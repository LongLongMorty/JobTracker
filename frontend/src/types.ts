export type Status =
    | 'WISHLIST'
    | 'APPLIED'
    | 'INTERVIEWING'
    | 'OFFERED'
    | 'REJECTED'
    | 'ARCHIVED';

export interface Application {
    id: number;
    company: string;
    position: string;
    status: Status;
    resume_version: string;
    jd_text: string;
    applied_at: string;
    location: string;
    salary_range: string;
    contact_info: string;
    notes: string;
    reached_interview: number;
    created_at: string;
    updated_at: string;
    interviews?: Interview[];
}

export interface ApplicationInput {
    company: string;
    position: string;
    resume_version: string;
    jd_text: string;
    applied_at: string;
    location: string;
    salary_range: string;
    contact_info: string;
    notes: string;
}

export interface Interview {
    id: number;
    application_id: number;
    round_name: string;
    scheduled_at: string;
    questions_and_notes: string;
    outcome: 'PENDING' | 'PASSED' | 'FAILED';
    created_at: string;
}

export interface InterviewInput {
    application_id: number;
    round_name: string;
    scheduled_at: string;
    questions_and_notes: string;
    outcome: 'PENDING' | 'PASSED' | 'FAILED' | '';
}

export interface Stats {
    total_applications: number;
    total_applied: number;
    interviewing: number;
    reached_interview: number;
    offered: number;
    rejected: number;
    archived: number;
    interview_rate: number;
    offer_rate: number;
    by_resume_version: ResumeVersionStat[];
}

export interface ResumeVersionStat {
    resume_version: string;
    total_applied: number;
    reached_interview: number;
    interview_rate: number;
}
