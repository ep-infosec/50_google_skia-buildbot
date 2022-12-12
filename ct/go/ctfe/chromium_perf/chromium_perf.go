/*
	Handlers and types specific to Chromium perf tasks.
*/

package chromium_perf

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"cloud.google.com/go/datastore"
	"github.com/gorilla/mux"
	"go.skia.org/infra/ct/go/ctfe/task_common"
	ctfeutil "go.skia.org/infra/ct/go/ctfe/util"
	ctutil "go.skia.org/infra/ct/go/util"
	"go.skia.org/infra/go/cas"
	"go.skia.org/infra/go/ds"
	"go.skia.org/infra/go/email"
	"go.skia.org/infra/go/swarming"
	skutil "go.skia.org/infra/go/util"
	"google.golang.org/api/iterator"
)

var (
	addTaskTemplate     *template.Template = nil
	runsHistoryTemplate *template.Template = nil
)

func ReloadTemplates(resourcesDir string) {
	addTaskTemplate = template.Must(template.ParseFiles(
		filepath.Join(resourcesDir, "chromium_perf.html"),
	))

	runsHistoryTemplate = template.Must(template.ParseFiles(
		filepath.Join(resourcesDir, "chromium_perf_runs.html"),
	))
}

type ChromiumPerfDatastoreTask struct {
	task_common.CommonCols

	Benchmark                    string   `json:"benchmark"`
	Platform                     string   `json:"platform"`
	RunOnGCE                     bool     `json:"run_on_gce"`
	PageSets                     string   `json:"page_sets"`
	IsTestPageSet                bool     `json:"is_test_page_set"`
	RepeatRuns                   int64    `json:"repeat_runs"`
	RunInParallel                bool     `json:"run_in_parallel"`
	GnArgs                       string   `json:"gn_args"`
	BenchmarkArgs                string   `json:"benchmark_args"`
	BrowserArgsNoPatch           string   `json:"browser_args_no_patch"`
	BrowserArgsWithPatch         string   `json:"browser_args_with_patch"`
	Description                  string   `json:"description"`
	CustomWebpagesGSPath         string   `json:"custom_webpages_gspath"`
	ChromiumPatchGSPath          string   `json:"chromium_patch_gspath"`
	BlinkPatchGSPath             string   `json:"blink_patch_gspath"`
	SkiaPatchGSPath              string   `json:"skia_patch_gspath"`
	CatapultPatchGSPath          string   `json:"catapult_patch_gspath"`
	BenchmarkPatchGSPath         string   `json:"benchmark_patch_gspath"`
	ChromiumPatchBaseBuildGSPath string   `json:"chromium_patch_base_build_gspath"`
	V8PatchGSPath                string   `json:"v8_patch_gspath"`
	Results                      string   `json:"results"`
	NoPatchRawOutput             string   `json:"no_patch_raw_output"`
	WithPatchRawOutput           string   `json:"with_patch_raw_output"`
	ChromiumHash                 string   `json:"chromium_hash"`
	CCList                       []string `json:"cc_list"`
	TaskPriority                 int      `json:"task_priority"`
	GroupName                    string   `json:"group_name"`
	ValueColumnName              string   `json:"value_column_name"`
}

func (task ChromiumPerfDatastoreTask) GetTaskName() string {
	return "ChromiumPerf"
}

func (task ChromiumPerfDatastoreTask) GetDescription() string {
	return task.Description
}

func (task ChromiumPerfDatastoreTask) GetPopulatedAddTaskVars() (task_common.AddTaskVars, error) {
	taskVars := &ChromiumPerfAddTaskVars{}
	taskVars.Username = task.Username
	taskVars.TsAdded = ctutil.GetCurrentTs()
	taskVars.RepeatAfterDays = strconv.FormatInt(task.RepeatAfterDays, 10)
	taskVars.Benchmark = task.Benchmark
	taskVars.Platform = task.Platform
	taskVars.RunOnGCE = strconv.FormatBool(task.RunOnGCE)
	taskVars.PageSets = task.PageSets
	taskVars.RepeatRuns = strconv.FormatInt(task.RepeatRuns, 10)
	taskVars.RunInParallel = strconv.FormatBool(task.RunInParallel)
	taskVars.GnArgs = task.GnArgs
	taskVars.BenchmarkArgs = task.BenchmarkArgs
	taskVars.BrowserArgsNoPatch = task.BrowserArgsNoPatch
	taskVars.BrowserArgsWithPatch = task.BrowserArgsWithPatch
	taskVars.Description = task.Description
	taskVars.ChromiumHash = task.ChromiumHash
	taskVars.CCList = task.CCList
	taskVars.TaskPriority = strconv.Itoa(task.TaskPriority)
	taskVars.GroupName = task.GroupName
	taskVars.ValueColumnName = task.ValueColumnName

	var err error
	taskVars.CustomWebpages, err = ctutil.GetPatchFromStorage(task.CustomWebpagesGSPath)
	if err != nil {
		return nil, fmt.Errorf("Could not read from %s: %s", task.CustomWebpagesGSPath, err)
	}
	taskVars.ChromiumPatch, err = ctutil.GetPatchFromStorage(task.ChromiumPatchGSPath)
	if err != nil {
		return nil, fmt.Errorf("Could not read from %s: %s", task.ChromiumPatchGSPath, err)
	}
	taskVars.BlinkPatch, err = ctutil.GetPatchFromStorage(task.BlinkPatchGSPath)
	if err != nil {
		return nil, fmt.Errorf("Could not read from %s: %s", task.BlinkPatchGSPath, err)
	}
	taskVars.SkiaPatch, err = ctutil.GetPatchFromStorage(task.SkiaPatchGSPath)
	if err != nil {
		return nil, fmt.Errorf("Could not read from %s: %s", task.SkiaPatchGSPath, err)
	}
	taskVars.CatapultPatch, err = ctutil.GetPatchFromStorage(task.CatapultPatchGSPath)
	if err != nil {
		return nil, fmt.Errorf("Could not read from %s: %s", task.CatapultPatchGSPath, err)
	}
	taskVars.BenchmarkPatch, err = ctutil.GetPatchFromStorage(task.BenchmarkPatchGSPath)
	if err != nil {
		return nil, fmt.Errorf("Could not read from %s: %s", task.BenchmarkPatchGSPath, err)
	}
	taskVars.V8Patch, err = ctutil.GetPatchFromStorage(task.V8PatchGSPath)
	if err != nil {
		return nil, fmt.Errorf("Could not read from %s: %s", task.V8PatchGSPath, err)
	}
	taskVars.ChromiumPatchBaseBuild, err = ctutil.GetPatchFromStorage(task.ChromiumPatchBaseBuildGSPath)
	if err != nil {
		return nil, fmt.Errorf("Could not read from %s: %s", task.ChromiumPatchBaseBuildGSPath, err)
	}

	return taskVars, nil
}

func (task ChromiumPerfDatastoreTask) GetResultsLink() string {
	return task.Results
}

func (task ChromiumPerfDatastoreTask) RunsOnGCEWorkers() bool {
	// Perf tasks should normally always run on bare-metal machines but we
	// also have Windows GCE instances now.
	return task.RunOnGCE
}

func (task ChromiumPerfDatastoreTask) GetDatastoreKind() ds.Kind {
	return ds.CHROMIUM_PERF_TASKS
}

func (task ChromiumPerfDatastoreTask) Query(it *datastore.Iterator) (interface{}, error) {
	tasks := []*ChromiumPerfDatastoreTask{}
	for {
		t := &ChromiumPerfDatastoreTask{}
		_, err := it.Next(t)
		if err == iterator.Done {
			break
		} else if err != nil {
			return nil, fmt.Errorf("Failed to retrieve list of tasks: %s", err)
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (task ChromiumPerfDatastoreTask) Get(c context.Context, key *datastore.Key) (task_common.Task, error) {
	t := &ChromiumPerfDatastoreTask{}
	if err := ds.DS.Get(c, key, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (task ChromiumPerfDatastoreTask) TriggerSwarmingTaskAndMail(ctx context.Context, swarmingClient swarming.ApiClient, casClient cas.CAS) error {
	runID := task_common.GetRunID(&task)
	emails := task_common.GetEmailRecipients(task.Username, task.CCList)
	cmd := []string{
		"cipd_bin_packages/luci-auth",
		"context",
		"--",
		"bin/run_chromium_perf_on_workers",
		"-logtostderr",
		"--run_requester=" + task.Username,
		"--description=" + task.Description,
		"--pageset_type=" + task.PageSets,
		"--gn_args=" + task.GnArgs,
		"--benchmark_name=" + task.Benchmark,
		"--benchmark_extra_args=" + task.BenchmarkArgs,
		"--browser_extra_args_nopatch=" + task.BrowserArgsNoPatch,
		"--browser_extra_args_withpatch=" + task.BrowserArgsWithPatch,
		"--repeat_benchmark=" + strconv.FormatInt(task.RepeatRuns, 10),
		"--run_in_parallel=" + strconv.FormatBool(task.RunInParallel),
		"--target_platform=" + task.Platform,
		"--run_on_gce=" + strconv.FormatBool(task.RunsOnGCEWorkers()),
		"--chromium_hash=" + task.ChromiumHash,
		"--run_id=" + runID,
		"--task_priority=" + strconv.Itoa(task.TaskPriority),
		"--group_name=" + task.GroupName,
		"--value_column_name=" + task.ValueColumnName,
		"--chromium_patch_gs_path=" + task.ChromiumPatchGSPath,
		"--skia_patch_gs_path=" + task.SkiaPatchGSPath,
		"--v8_patch_gs_path=" + task.V8PatchGSPath,
		"--catapult_patch_gs_path=" + task.CatapultPatchGSPath,
		"--chromium_base_build_patch_gs_path=" + task.ChromiumPatchBaseBuildGSPath,
		"--custom_webpages_csv_gs_path=" + task.CustomWebpagesGSPath,
	}
	sTaskID, err := ctutil.TriggerMasterScriptSwarmingTask(ctx, runID, "run_chromium_perf_on_workers", false, cmd, ctutil.CasChromiumPerfMaster(), swarmingClient, casClient)
	if err != nil {
		return fmt.Errorf("Could not trigger master script for run_chromium_perf_on_workers with cmd %v: %s", cmd, err)
	}
	// Mark task as started in datastore.
	if err := task_common.UpdateTaskSetStarted(ctx, runID, sTaskID, &task); err != nil {
		return fmt.Errorf("Could not mark task as started in datastore: %s", err)
	}
	// Send start email.
	skutil.LogErr(ctfeutil.SendTaskStartEmail(task.DatastoreKey.ID, emails, "Chromium perf", runID, task.Description, fmt.Sprintf("Triggered %s benchmark on %s %s pageset.", task.Benchmark, task.Platform, task.PageSets)))

	return nil
}

func (task ChromiumPerfDatastoreTask) SendCompletionEmail(ctx context.Context, completedSuccessfully bool) error {
	runID := task_common.GetRunID(&task)
	emails := task_common.GetEmailRecipients(task.Username, task.CCList)
	emailSubject := fmt.Sprintf("Cluster telemetry chromium perf task has completed (#%d)", task.DatastoreKey.ID)
	failureHtml := ""
	viewActionMarkup := ""
	ctPerfHtml := ""
	var err error

	if completedSuccessfully {
		if viewActionMarkup, err = email.GetViewActionMarkup(task.Results, "View Results", "Direct link to the HTML results"); err != nil {
			return fmt.Errorf("Failed to get view action markup: %s", err)
		}
		ctPerfHtml = ctfeutil.GetCTPerfEmailHtml(task.GroupName)
	} else {
		emailSubject += " with failures"
		failureHtml = ctfeutil.GetFailureEmailHtml(runID)
		if viewActionMarkup, err = email.GetViewActionMarkup(fmt.Sprintf(ctutil.SWARMING_RUN_ID_ALL_TASKS_LINK_TEMPLATE, runID), "View Failure", "Direct link to the swarming logs"); err != nil {
			return fmt.Errorf("Failed to get view action markup: %s", err)
		}
	}
	bodyTemplate := `
	The chromium perf %s benchmark task on %s pageset has completed. %s.<br/>
	Run description: %s<br/>
	%s
	%s
	The HTML output with differences between the base run and the patch run is <a href='%s'>here</a>.<br/>
	The patch(es) you specified are here:
	<a href='%s'>chromium</a>/<a href='%s'>skia</a>/<a href='%s'>v8</a>/<a href='%s'>catapult</a>/<a href='%s'>chromium (base build)</a>
	<br/>
	Custom webpages (if specified) are <a href='%s'>here</a>.
	<br/><br/>
	You can schedule more runs <a href='%s'>here</a>.
	<br/><br/>
	Thanks!
	`
	chromiumPatchLink := ctutil.GCS_HTTP_LINK + path.Join(ctutil.GCSBucketName, task.ChromiumPatchGSPath)
	skiaPatchLink := ctutil.GCS_HTTP_LINK + path.Join(ctutil.GCSBucketName, task.SkiaPatchGSPath)
	v8PatchLink := ctutil.GCS_HTTP_LINK + path.Join(ctutil.GCSBucketName, task.V8PatchGSPath)
	catapultPatchLink := ctutil.GCS_HTTP_LINK + path.Join(ctutil.GCSBucketName, task.CatapultPatchGSPath)
	chromiumPatchBaseBuildLink := ctutil.GCS_HTTP_LINK + path.Join(ctutil.GCSBucketName, task.ChromiumPatchBaseBuildGSPath)
	customWebpagesLink := ctutil.GCS_HTTP_LINK + path.Join(ctutil.GCSBucketName, task.CustomWebpagesGSPath)
	emailBody := fmt.Sprintf(bodyTemplate, task.Benchmark, task.PageSets, ctfeutil.GetSwarmingLogsLink(runID), task.Description, failureHtml, ctPerfHtml, task.Results, chromiumPatchLink, skiaPatchLink, v8PatchLink, catapultPatchLink, chromiumPatchBaseBuildLink, customWebpagesLink, task_common.WebappURL+ctfeutil.CHROMIUM_PERF_URI)
	if err := ctfeutil.SendEmailWithMarkup(emails, emailSubject, emailBody, viewActionMarkup); err != nil {
		return fmt.Errorf("Error while sending email: %s", err)
	}
	return nil
}

func (task *ChromiumPerfDatastoreTask) SetCompleted(success bool) {
	if success {
		runID := task_common.GetRunID(task)
		task.Results = ctutil.GetPerfOutputLink(runID)
		task.NoPatchRawOutput = ctutil.GetPerfNoPatchOutputLink(runID)
		task.WithPatchRawOutput = ctutil.GetPerfWithPatchOutputLink(runID)
	}
	task.TsCompleted = ctutil.GetCurrentTsInt64()
	task.Failure = !success
	task.TaskDone = true
}

func addTaskView(w http.ResponseWriter, r *http.Request) {
	ctfeutil.ExecuteSimpleTemplate(addTaskTemplate, w, r)
}

type ChromiumPerfAddTaskVars struct {
	task_common.AddTaskCommonVars

	Benchmark            string   `json:"benchmark"`
	Platform             string   `json:"platform"`
	RunOnGCE             string   `json:"run_on_gce"`
	PageSets             string   `json:"page_sets"`
	CustomWebpages       string   `json:"custom_webpages"`
	RepeatRuns           string   `json:"repeat_runs"`
	RunInParallel        string   `json:"run_in_parallel"`
	GnArgs               string   `json:"gn_args"`
	BenchmarkArgs        string   `json:"benchmark_args"`
	BrowserArgsNoPatch   string   `json:"browser_args_nopatch"`
	BrowserArgsWithPatch string   `json:"browser_args_withpatch"`
	Description          string   `json:"desc"`
	ChromiumHash         string   `json:"chromium_hash"`
	CCList               []string `json:"cc_list"`
	TaskPriority         string   `json:"task_priority"`
	GroupName            string   `json:"group_name"`
	ValueColumnName      string   `json:"value_column_name"`

	ChromiumPatch          string `json:"chromium_patch"`
	BlinkPatch             string `json:"blink_patch"`
	SkiaPatch              string `json:"skia_patch"`
	CatapultPatch          string `json:"catapult_patch"`
	BenchmarkPatch         string `json:"benchmark_patch"`
	V8Patch                string `json:"v8_patch"`
	ChromiumPatchBaseBuild string `json:"chromium_patch_base_build"`
}

func (task *ChromiumPerfAddTaskVars) GetDatastoreKind() ds.Kind {
	return ds.CHROMIUM_PERF_TASKS
}

func (task *ChromiumPerfAddTaskVars) GetPopulatedDatastoreTask(ctx context.Context) (task_common.Task, error) {
	if task.Benchmark == "" ||
		task.Platform == "" ||
		task.PageSets == "" ||
		task.RepeatRuns == "" ||
		task.RunInParallel == "" ||
		task.Description == "" {
		return nil, fmt.Errorf("Invalid parameters")
	}
	if task.GroupName != "" && len(task.GroupName) >= ctfeutil.MAX_GROUPNAME_LEN {
		return nil, fmt.Errorf("Please limit group names to less than %d characters", ctfeutil.MAX_GROUPNAME_LEN)
	}

	customWebpagesSlice, err := ctfeutil.GetQualifiedCustomWebpages(task.CustomWebpages, task.BenchmarkArgs)
	if err != nil {
		return nil, err
	}
	customWebpages := strings.Join(customWebpagesSlice, ",")
	customWebpagesGSPath, err := ctutil.SavePatchToStorage(customWebpages)
	if err != nil {
		return nil, fmt.Errorf("Could not save custom webpages to storage: %s", err)
	}

	chromiumPatchGSPath, err := ctutil.SavePatchToStorage(task.ChromiumPatch)
	if err != nil {
		return nil, fmt.Errorf("Could not save chromium patch to storage: %s", err)
	}
	blinkPatchGSPath, err := ctutil.SavePatchToStorage(task.BlinkPatch)
	if err != nil {
		return nil, fmt.Errorf("Could not save blink patch to storage: %s", err)
	}
	skiaPatchGSPath, err := ctutil.SavePatchToStorage(task.SkiaPatch)
	if err != nil {
		return nil, fmt.Errorf("Could not save skia patch to storage: %s", err)
	}
	catapultPatchGSPath, err := ctutil.SavePatchToStorage(task.CatapultPatch)
	if err != nil {
		return nil, fmt.Errorf("Could not save catapult patch to storage: %s", err)
	}
	benchmarkPatchGSPath, err := ctutil.SavePatchToStorage(task.BenchmarkPatch)
	if err != nil {
		return nil, fmt.Errorf("Could not save benchmark patch to storage: %s", err)
	}
	v8PatchGSPath, err := ctutil.SavePatchToStorage(task.V8Patch)
	if err != nil {
		return nil, fmt.Errorf("Could not save v8 patch to storage: %s", err)
	}
	chromiumPatchBaseBuildGSPath, err := ctutil.SavePatchToStorage(task.ChromiumPatchBaseBuild)
	if err != nil {
		return nil, fmt.Errorf("Could not save chromium patch for base build to storage: %s", err)
	}

	t := &ChromiumPerfDatastoreTask{
		Benchmark:            task.Benchmark,
		Platform:             task.Platform,
		PageSets:             task.PageSets,
		IsTestPageSet:        task.PageSets == ctutil.PAGESET_TYPE_DUMMY_1k || task.PageSets == ctutil.PAGESET_TYPE_MOBILE_DUMMY_1k,
		GnArgs:               task.GnArgs,
		BenchmarkArgs:        task.BenchmarkArgs,
		BrowserArgsNoPatch:   task.BrowserArgsNoPatch,
		BrowserArgsWithPatch: task.BrowserArgsWithPatch,
		Description:          task.Description,
		ChromiumHash:         task.ChromiumHash,
		CCList:               task.CCList,
		GroupName:            task.GroupName,
		ValueColumnName:      task.ValueColumnName,

		CustomWebpagesGSPath:         customWebpagesGSPath,
		ChromiumPatchGSPath:          chromiumPatchGSPath,
		BlinkPatchGSPath:             blinkPatchGSPath,
		SkiaPatchGSPath:              skiaPatchGSPath,
		CatapultPatchGSPath:          catapultPatchGSPath,
		BenchmarkPatchGSPath:         benchmarkPatchGSPath,
		V8PatchGSPath:                v8PatchGSPath,
		ChromiumPatchBaseBuildGSPath: chromiumPatchBaseBuildGSPath,
	}
	runOnGCE, err := strconv.ParseBool(task.RunOnGCE)
	if err != nil {
		return nil, fmt.Errorf("%s is not bool: %s", task.RunOnGCE, err)
	}
	t.RunOnGCE = runOnGCE
	runInParallel, err := strconv.ParseBool(task.RunInParallel)
	if err != nil {
		return nil, fmt.Errorf("%s is not bool: %s", task.RunInParallel, err)
	}
	t.RunInParallel = runInParallel
	repeatRuns, err := strconv.ParseInt(task.RepeatRuns, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%s is not int64: %s", task.RepeatRuns, err)
	}
	t.RepeatRuns = repeatRuns
	taskPriority, err := strconv.Atoi(task.TaskPriority)
	if err != nil {
		return nil, fmt.Errorf("%s is not int: %s", task.TaskPriority, err)
	}
	if taskPriority == 0 {
		// This should only happen for repeating tasks that were created before
		// support for task priorities was added to CT.
		// Triggering tasks with 0 priority fails in swarming with
		// "priority 0 can only be used for terminate request"
		// Override it to the medium priority.
		taskPriority = ctutil.TASKS_PRIORITY_MEDIUM
	}
	t.TaskPriority = taskPriority
	return t, nil
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	task_common.AddTaskHandler(w, r, &ChromiumPerfAddTaskVars{})
}

func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	task_common.GetTasksHandler(&ChromiumPerfDatastoreTask{}, w, r)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	task_common.DeleteTaskHandler(&ChromiumPerfDatastoreTask{}, w, r)
}

func redoTaskHandler(w http.ResponseWriter, r *http.Request) {
	task_common.RedoTaskHandler(&ChromiumPerfDatastoreTask{}, w, r)
}

func editTaskHandler(w http.ResponseWriter, r *http.Request) {
	task_common.EditTaskHandler(&ChromiumPerfDatastoreTask{}, w, r)
}

func runsHistoryView(w http.ResponseWriter, r *http.Request) {
	ctfeutil.ExecuteSimpleTemplate(runsHistoryTemplate, w, r)
}

func AddHandlers(externalRouter *mux.Router) {
	externalRouter.HandleFunc("/", addTaskView).Methods("GET")
	externalRouter.HandleFunc("/"+ctfeutil.CHROMIUM_PERF_URI, addTaskView).Methods("GET")
	externalRouter.HandleFunc("/"+ctfeutil.CHROMIUM_PERF_RUNS_URI, runsHistoryView).Methods("GET")

	externalRouter.HandleFunc("/"+ctfeutil.ADD_CHROMIUM_PERF_TASK_POST_URI, addTaskHandler).Methods("POST")
	externalRouter.HandleFunc("/"+ctfeutil.GET_CHROMIUM_PERF_TASKS_POST_URI, getTasksHandler).Methods("POST")
	externalRouter.HandleFunc("/"+ctfeutil.DELETE_CHROMIUM_PERF_TASK_POST_URI, deleteTaskHandler).Methods("POST")
	externalRouter.HandleFunc("/"+ctfeutil.REDO_CHROMIUM_PERF_TASK_POST_URI, redoTaskHandler).Methods("POST")
	externalRouter.HandleFunc("/"+ctfeutil.EDIT_CHROMIUM_PERF_TASK_POST_URI, editTaskHandler).Methods("POST")
}
