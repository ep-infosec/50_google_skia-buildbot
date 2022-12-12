// DO NOT EDIT. This file is automatically generated.

export namespace pivot {
	export interface Request {
		group_by: string[] | null;
		operation: pivot.Operation;
		summary: pivot.Operation[] | null;
	}
}

export interface Go2TS {
}

export interface Alert {
	id_as_string: string;
	display_name: string;
	query: string;
	alert: string;
	interesting: number;
	bug_uri_template: string;
	algo: ClusterAlgo;
	step: StepDetection;
	state: ConfigState;
	owner: string;
	step_up_only: boolean;
	direction: Direction;
	radius: number;
	k: number;
	group_by: string;
	sparse: boolean;
	minimum_num: number;
	category: string;
}

export interface AlertsStatus {
	alerts: number;
}

export interface ValuePercent {
	value: string;
	percent: number;
}

export interface StepFit {
	least_squares: number;
	turning_point: number;
	step_size: number;
	regression: number;
	status: StepFitStatus;
}

export interface ColumnHeader {
	offset: CommitNumber;
	timestamp: number;
}

export interface ClusterSummary {
	centroid: number[] | null;
	shortcut: string;
	param_summaries2: ValuePercent[] | null;
	step_fit: StepFit | null;
	step_point: ColumnHeader | null;
	num: number;
	ts: string;
}

export interface Commit {
	offset: CommitNumber;
	hash: string;
	ts: number;
	author: string;
	message: string;
	url: string;
}

export interface Current {
	commit: Commit;
	alert: Alert | null;
	message: string;
}

export interface DataFrame {
	traceset: TraceSet;
	header: (ColumnHeader | null)[] | null;
	paramset: ReadOnlyParamSet;
	skip: number;
}

export interface FrameResponse {
	dataframe: DataFrame | null;
	skps: number[] | null;
	msg: string;
	display_mode: FrameResponseDisplayMode;
}

export interface TriageStatus {
	status: Status;
	message: string;
}

export interface Regression {
	low: ClusterSummary | null;
	high: ClusterSummary | null;
	frame: FrameResponse | null;
	low_status: TriageStatus;
	high_status: TriageStatus;
}

export interface RegressionAtCommit {
	cid: Commit;
	regression: Regression | null;
}

export interface FrameRequest {
	begin: number;
	end: number;
	formulas: string[] | null;
	queries: string[] | null;
	keys: string;
	tz: string;
	num_commits: number;
	request_type: RequestType;
	pivot: pivot.Request | null;
}

export interface AlertUpdateResponse {
	IDAsString: string;
}

export interface CIDHandlerResponse {
	commitSlice: Commit[] | null;
	logEntry: string;
}

export interface ClusterStartResponse {
	id: string;
}

export interface CommitDetailsRequest {
	cid: CommitNumber;
	traceid: string;
}

export interface CountHandlerRequest {
	q: string;
	begin: number;
	end: number;
}

export interface CountHandlerResponse {
	count: number;
	paramset: ReadOnlyParamSet;
}

export interface RangeRequest {
	offset: CommitNumber;
	begin: number;
	end: number;
}

export interface RegressionRangeRequest {
	begin: number;
	end: number;
	subset: Subset;
	alert_filter: string;
}

export interface RegressionRow {
	cid: Commit;
	columns: (Regression | null)[] | null;
}

export interface RegressionRangeResponse {
	header: (Alert | null)[] | null;
	table: (RegressionRow | null)[] | null;
	categories: string[] | null;
}

export interface ShiftRequest {
	begin: CommitNumber;
	end: CommitNumber;
}

export interface ShiftResponse {
	begin: number;
	end: number;
}

export interface SkPerfConfig {
	radius: number;
	key_order: string[] | null;
	num_shift: number;
	interesting: number;
	step_up_only: boolean;
	commit_range_url: string;
	demo: boolean;
	display_group_by: boolean;
}

export interface TriageRequest {
	cid: CommitNumber;
	alert: Alert;
	triage: TriageStatus;
	cluster_type: string;
}

export interface TriageResponse {
	bug: string;
}

export interface TryBugRequest {
	bug_uri_template: string;
}

export interface TryBugResponse {
	url: string;
}

export interface FullSummary {
	summary: ClusterSummary;
	triage: TriageStatus;
	frame: FrameResponse;
}

export interface Domain {
	n: number;
	end: string;
	offset: number;
}

export interface RegressionDetectionRequest {
	alert: Alert | null;
	domain: Domain;
	step: number;
	total_queries: number;
}

export interface ClusterSummaries {
	Clusters: (ClusterSummary | null)[] | null;
	StdDevThreshold: number;
	K: number;
}

export interface RegressionDetectionResponse {
	summary: ClusterSummaries | null;
	frame: FrameResponse | null;
}

export interface TryBotRequest {
	kind: TryBotRequestKind;
	cl: CL;
	patch_number: number;
	commit_number: CommitNumber;
	query: string;
}

export interface TryBotResult {
	params: Params;
	median: number;
	lower: number;
	upper: number;
	stddevRatio: number;
	values: number[] | null;
}

export interface TryBotResponse {
	header: (ColumnHeader | null)[] | null;
	results: TryBotResult[] | null;
	paramset: ReadOnlyParamSet;
}

export namespace progress {
	export interface Message {
		key: string;
		value: string;
	}
}

export namespace progress {
	export interface SerializedProgress {
		status: progress.Status;
		messages: progress.Message[];
		results?: any;
		url: string;
	}
}

export namespace ingest {
	export interface SingleMeasurement {
		value: string;
		measurement: number;
	}
}

export namespace ingest {
	export interface Result {
		key: { [key: string]: string } | null;
		measurement?: number;
		measurements?: { [key: string]: ingest.SingleMeasurement[] | null } | null;
	}
}

export namespace ingest {
	export interface Format {
		version: number;
		git_hash: string;
		issue?: CL;
		patchset?: string;
		key?: { [key: string]: string } | null;
		results: ingest.Result[] | null;
		links?: { [key: string]: string } | null;
	}
}

export type Params = { [key: string]: string };

export type ParamSet = { [key: string]: string[] };

export type ReadOnlyParamSet = { [key: string]: string[] };

export type Trace = number[];

export type TraceSet = { [key: string]: Trace };

export namespace pivot { export type Operation = "sum" | "avg" | "geo" | "std" | "count" | "min" | "max"; }

export type ClusterAlgo = "kmeans" | "stepfit";

export type StepDetection = "" | "absolute" | "const" | "percent" | "cohen" | "mannwhitneyu";

export type ConfigState = "ACTIVE" | "DELETED";

export type Direction = "UP" | "DOWN" | "BOTH";

export type StepFitStatus = "Low" | "High" | "Uninteresting";

export type CommitNumber = number;

export type FrameResponseDisplayMode = "display_query_only" | "display_plot" | "display_pivot_table" | "display_pivot_plot" | "display_spinner";

export type Status = "" | "positive" | "negative" | "untriaged";

export type RequestType = 0 | 1;

export type Subset = "all" | "regressions" | "untriaged";

export type TryBotRequestKind = "trybot" | "commit";

export type CL = string;

export type ProcessState = "Running" | "Success" | "Error";

export namespace progress { export type Status = "Running" | "Finished" | "Error"; }