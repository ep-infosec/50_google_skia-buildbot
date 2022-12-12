// DO NOT EDIT. This file is automatically generated.

export interface IssueCountsData {
	open_count: number;
	unassigned_count: number;
	untriaged_count: number;
	p0_count: number;
	p1_count: number;
	p2_count: number;
	p3_count: number;
	p4_count: number;
	p5_count: number;
	p6_count: number;
	p0_slo_count: number;
	p1_slo_count: number;
	p2_slo_count: number;
	p3_slo_count: number;
	query_link: string;
	untriaged_query_link: string;
	p0_link: string;
	p1_link: string;
	p2_link: string;
	p3_and_rest_link: string;
}

export interface Issue {
	id: string;
	state: string;
	priority: StandardizedPriority;
	owner: string;
	link: string;
	slo_violation: boolean;
	slo_violation_reason: string;
	slo_violation_duration: Duration;
	created: string;
	modified: string;
	title: string;
	summary: string;
}

export interface ClientSourceQueryRequest {
	client: RecognizedClient;
	source: IssueSource;
	query: string;
}

export interface IssuesOutsideSLOResponse {
	pri_to_slo_issues: { [key: string]: (Issue | null)[] | null } | null;
}

export interface GetClientsResponse {
	clients: { [key: string]: { [key: string]: { [key: string]: boolean } | null } | null } | null;
}

export interface StatusData {
	untriaged_count: number;
	link: string;
}

export interface GetClientCountsResponse {
	clients_to_status_data: { [key: string]: StatusData } | null;
}

export interface GetChartsDataResponse {
	open_data: any;
	slo_data: any;
	untriaged_data: any;
}

export type StandardizedPriority = string;

export type Duration = number;

export type RecognizedClient = string;

export type IssueSource = string;
