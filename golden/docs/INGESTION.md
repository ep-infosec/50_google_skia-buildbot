Ingesting data into Gold
========================

Overview
--------

Gold ingests images and JSON files. The images are also referred to as
'digests' since we rely on a content addressable approach where the file name
of the images is generally the (MD5) hash of the image.

Note: For Skia we do not hash the ingested PNG file, but the internal Skia bitmap of
the image. The desired property of content addressability remains because
identical digests refer to identical images.

Each ingested JSON file describes a single bot run for a specific commit.
It generally does not refer to images directly - it only refers to their digests.
The names of images are then derived from the digests.

Both images and JSON files are stored in GCS (Google Cloud Storage) and then ingested
by the Gold process (running in GCE).
Generally a process will want to write the images of a bot run first.
Only if the images have been uploaded to GCS successfully, the JSON file should be
added to GCS. The content in GCS is considered the 'source for truth' for Gold.

Note: Since all images are content addressable - we only need to upload images
that are not already in GCS.

Storage Layout in GCS
--------------------

JSON files are stored at

    gs://JSON_BUCKET/JSON_DIR/YYYY/MM/DD/HH/GIT_HASH/BUILDER_NAME/BUILD_NUMBER/dm.json

Where JSON_BUCKET and JSON_DIR are the GCS bucket and directory respectively.
YYYY, MM, DD and HH are the year, month, day and hour (0-23) respectively of
when the bot run finished. All times are based on UTC.
GIT_HASH is the value of the git commit hash, BUILDER_NAME and BUILDER_NUMBER
refer to the bot instance and run that produced the output.

Here is an example of a valid uploaded JSON file (requires permissions to the bucket):

gs://skia-infra-gm/dm-json-v1/2014/09/17/15/4aa6dfc0b77af9ac298bb9d48991b72a2fec00b2/Test-Android-Xoom-Tegra2-Arm7-Release/3056/dm.json

Images are stored at

    gs://IMAGE_BUCKET/IMAGE_DIR/<<DIGEST>>.png

Where IMAGE_BUCKET and IMAGE_DIR are the bucket and directory under which all
images are stored.

<<DIGEST>> is the digest generated from the image content and used to refer to
the image by the JSON file.

Note: Most information encoded in the path is also contained in the JSON file
itself. The path information is used by Gold ingestion to scan for new files
continuously. So it's important that the date in the path is the actual date of
when the data were generated and it has to be based on the UTC timezone.

The bucket and directory values for JSON files and images are shared between the
bot and the Gold ingestion process.

JSON Input file
---------------

The JSON file intended to be simple with  flexibility for the specific application
that generates the baseline images.
(See below for a tool to validate JSON input to Gold.)

Here is a shortened but representative example of the input format:
```json5
{
   "gitHash" : "c4711517219f333c1116f47706eb57b51b5f8fc7",
   "key" : {
      "arch" : "arm64",
      "compiler" : "Clang",
      "configuration" : "Debug",
      "cpu_or_gpu" : "GPU",
      "cpu_or_gpu_value" : "PowerVRGT7600",
      "extra_config" : "Metal",
      "model" : "iPhone7",
      "os" : "iOS"
   },
   "results" : [
      {
         "key" : {
            "config" : "mtl",
            "name" : "yuv_nv12_to_rgb_effect",
            "source_type" : "gm"
         },
         "md5" : "30a470b6ac174aa1ffb54fcb77a21f21",
         "options" : {
            "ext" : "png",
            "gamma_correct" : "no"
         }
      },
      {
         "key" : {
            "config" : "mtl",
            "name" : "yuv_to_rgb_effect",
            "source_type" : "gm"
         },
         "md5" : "0ea32027e1e651e4250797aa44bfadaa",
         "options" : {
            "ext" : "png",
            "gamma_correct" : "no"
         }
      },
      {
         "key" : {
            "config" : "pipe-8888",
            "name" : "clipcubic",
            "source_type" : "gm"
         },
         "md5" : "64e446d96bebba035887dd7dda6db6c4",
         "options" : {
            "ext" : "png"
         }
      }
   ],
   // These keys are required for tryjobs and can be omitted for non-tryjobs.
   // GitHub support coming soon, Gerrit/googlesource support only at the moment.
   "issue": "0",
   "patchset": "0",
   "buildbucket_build_id" : "0",
   // These keys are optional, but can assist in debugging
   "builder" : "Test-Android-Clang-iPhone7-GPU-PowerVRGT7600-arm64-Debug-All-Metal",
   "swarming_bot_id" : "skia-rpi-102",
   "swarming_task_id" : "3fcd8d4a539ba311",
}
```

In the root of the object these fields are required:

* gitHash: The git commit hash of the version being tested (Not important
  for trybot runs).

* issue: Only relevant for trybot runs (before a code change is commited). It
  refers to the Gerrit issue that contains the change list being tested.

* patchset: Only relevant for trybot runs. It refers to the patchset within the
  issue that was used for this test run.

* key: The set of key-value pairs shared by all results. This is usually the
  hardware/OS configuration of the bot that ran the test. These are
  application dependent key-value pairs and are used later by Gold's UI to
  filter results.

* results: A list of results each representing an image that generated by a
  specific test and configuration.
  The objects in 'result' need to contain at least the following fields:

    - key.name: the name of the test.

    - key.source_type: used to group different tests together. This has to be
      present even if you don't have different test groups. In that case
      simply use a constant value.

    - md5: the digest of the resulting image. It does not have to be MD5 based,
      but should be a hash (with MD5 like properties) that is unique to the
      resulting image. This is used by Gold later to fetch the images associated
      with the test.

    - options.ext: The file type. This needs to be "png" for the test to be
      ingested.

 * options: these keys are meant as an FYI - they can be filtered by, but they
   do not impact the trace uniqueness.

Validating Gold input with goldctl
----------------------------------

To validate whether JSON is valid Gold input you can use the `goldctl` tool.

To validate a JSON file run one of these:

```console
   $ goldctl validate -f dm.json
   $ cat dm.json | goldctl validate
```

A successful run returns an exit code of zero, but produces no output.
If there are issues, expect to see something like this:

```console
   $ goldctl validate -f dm.json
     JSON validation failed:
       field 'gitHash' must be hexadecimal. Received ''
     exit status 1
   $
```

Running

```console
   $ goldctl help validate
```

will output basic information about how to use the validate command.
