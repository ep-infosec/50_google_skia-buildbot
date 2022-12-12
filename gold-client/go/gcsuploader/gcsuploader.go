package gcsuploader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
	"time"

	gstorage "cloud.google.com/go/storage"
	"github.com/cenkalti/backoff"
	"google.golang.org/api/option"

	"go.skia.org/infra/go/exec"
	"go.skia.org/infra/go/gcs"
	"go.skia.org/infra/go/skerr"
)

const (
	// gcsPrefix is the expected prefix for a GCS URL.
	gcsPrefix = "gs://"
)

// GCSUploader implementations provide functions to upload to GCS.
type GCSUploader interface {
	// UploadBytes copies a local file to GCS. If data is provided, those
	// bytes may be used instead of read again from disk.
	// The dst string is assumed to have a gs:// prefix.
	// Currently only uploading from a local file to GCS is supported, that is
	// one cannot use gs://foo/bar as 'fileName'
	UploadBytes(ctx context.Context, data []byte, fallbackSrc, dst string) error

	// UploadJSON serializes the given data to JSON and uploads the result to GCS.
	// An implementation can use tempFileName for temporary storage of JSON data.
	UploadJSON(ctx context.Context, data interface{}, tempFileName, gcsObjectPath string) error
}

// GsutilImpl implements the  GCSUploader and ImageDownloader interfaces.
type GsutilImpl struct{}

// UploadJSON serializes the given data to JSON and writes the result to the given
// tempFileName, then it copies the file to the given path in GCS. gcsObjPath is assumed
// to have the form: <bucket_name>/path/to/object
func (g *GsutilImpl) UploadJSON(ctx context.Context, data interface{}, tempFileName, gcsObjPath string) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return skerr.Wrapf(err, "could not marshal to JSON before uploading")
	}

	if err := ioutil.WriteFile(tempFileName, jsonBytes, 0644); err != nil {
		return skerr.Wrapf(err, "saving json to %s", tempFileName)
	}

	// Upload the written file.
	return g.UploadBytes(ctx, nil, tempFileName, prefixGCS(gcsObjPath))
}

// prefixGCS adds the "gs://" prefix to the given GCS path.
func prefixGCS(gcsPath string) string {
	return fmt.Sprintf(gcsPrefix+"%s", gcsPath)
}

// UploadBytes shells out to gsutil to copy the given src to the given target. A path
// starting with "gs://" is assumed to be in GCS.
func (g *GsutilImpl) UploadBytes(ctx context.Context, _ []byte, fallbackSrc, dst string) error {
	return g.gsutilCmd(ctx, "cp", fallbackSrc, dst)
}

// gsutilCmd executes a given command using the local gsutil executable (or python script, if
// on Windows).
func (g *GsutilImpl) gsutilCmd(ctx context.Context, cmd ...string) error {
	var outBuf bytes.Buffer
	runCmd := &exec.Command{
		Name:           "gsutil",
		Args:           cmd,
		CombinedOutput: &outBuf,
	}
	if err := exec.Run(ctx, runCmd); err != nil {
		if runtime.GOOS == "windows" {
			cmd = append([]string{"gsutil.py"}, cmd...)
			runCmd = &exec.Command{
				Name:           "python",
				Args:           cmd,
				CombinedOutput: &outBuf,
			}
			if err := exec.Run(ctx, runCmd); err != nil {
				return skerr.Wrapf(err, "running gsutil on windows. Got output \n%s\n", outBuf.String())
			}
		} else {
			return skerr.Wrapf(err, "running gsutil. Got output \n%s\n", outBuf.String())
		}
	}
	return nil
}

// clientImpl implements the GCSUploader interface using an authenticated
// (via an OAuth service account) http client.
type clientImpl struct {
	client *gstorage.Client
}

func New(ctx context.Context, httpClient *http.Client) (*clientImpl, error) {
	ret := &clientImpl{}
	var err error
	ret.client, err = gstorage.NewClient(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, skerr.Wrapf(err, "instantiating storage client")
	}
	return ret, nil
}

// UploadBytes implements the GCSUploader interface.
func (h *clientImpl) UploadBytes(ctx context.Context, data []byte, fallbackSrc, dst string) error {
	if len(data) == 0 {
		if strings.HasPrefix(fallbackSrc, gcsPrefix) {
			return skerr.Fmt("Copying from a remote file is not supported")
		}

		var err error
		data, err = ioutil.ReadFile(fallbackSrc)
		if err != nil {
			return skerr.Wrapf(err, "reading file %s", fallbackSrc)
		}
	}

	return h.uploadToGCS(ctx, data, dst)
}

// UploadJSON implements the GCSUploader interface.
func (h *clientImpl) UploadJSON(ctx context.Context, data interface{}, _, gcsObjectPath string) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return skerr.Wrap(err)
	}
	return h.uploadToGCS(ctx, jsonBytes, gcsObjectPath)
}

// uploadToGCS takes the given bytes and uploads them to the destination GCS object.
func (h *clientImpl) uploadToGCS(ctx context.Context, data []byte, dst string) error {
	// Trim the prefix and upload the content to the cloud.
	dst = strings.TrimPrefix(dst, gcsPrefix)
	bucket, objPath := gcs.SplitGSPath(dst)
	handle := h.client.Bucket(bucket).Object(objPath)

	isFirstTime := true
	eb := backoff.NewExponentialBackOff()
	eb.InitialInterval = time.Second
	eb.MaxInterval = 10 * time.Second
	eb.MaxElapsedTime = 30 * time.Second
	err := backoff.Retry(func() error {
		if err := ctx.Err(); err != nil {
			return backoff.Permanent(err)
		}
		// To avoid slowing down every upload with an extra check, we only check if the file
		// already exists on the first failure (e.g. 503 that happens if multiple processes are
		// writing the same file at once).
		if !isFirstTime {
			if _, err := handle.Attrs(ctx); err == nil {
				fmt.Printf("Skipping upload of %s - already exists on server\n", dst)
				return nil
			}
		}
		isFirstTime = false
		ctx, cancel := context.WithCancel(ctx)
		defer cancel() // The docs say to cancel the context in the event of an error or success.
		w := handle.NewWriter(ctx)
		w.ChunkSize = 0 // Upload files in one go, to avoid conflicts if multiple processes are
		// uploading at the same time.
		_, err := w.Write(data)
		if err != nil {
			fmt.Printf("Error writing to %s: %s (retrying)\n", dst, err)
			return skerr.Wrap(err)
		}
		if err := w.Close(); err != nil {
			fmt.Printf("Error closing file %s: %s (retrying)\n", dst, err)
			return skerr.Wrap(err)
		}
		return nil
	}, eb)
	if err != nil {
		return skerr.Wrapf(err, "uploading %d bytes to %s with exponential retries", len(data), dst)
	}
	return nil
}

// DryRunImpl implements the GCSUploader and ImageDownloader interfaces (but doesn't
// actually upload or download anything)
type DryRunImpl struct{}

// UploadBytes implements the GCSUploader interface.
func (h *DryRunImpl) UploadBytes(_ context.Context, _ []byte, fallbackSrc, dst string) error {
	fmt.Printf("dryrun -- upload bytes from %s to %s\n", fallbackSrc, dst)
	return nil
}

// UploadJSON implements the GCSUploader interface.
func (h *DryRunImpl) UploadJSON(_ context.Context, _ interface{}, tempFileName, gcsObjectPath string) error {
	fmt.Printf("dryrun -- upload JSON from %s to %s\n", tempFileName, gcsObjectPath)
	return nil
}

// Make sure GsutilImpl fulfills the GCSUploader interface.
var _ GCSUploader = (*GsutilImpl)(nil)

// Make sure DryRunImpl fulfills the GCSUploader interface.
var _ GCSUploader = (*DryRunImpl)(nil)
