export namespace main {
	
	export class QAItem {
	    id: number;
	    interview_id: number;
	    question: string;
	    answer: string;
	    reflection: string;
	    sort_order: number;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new QAItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.interview_id = source["interview_id"];
	        this.question = source["question"];
	        this.answer = source["answer"];
	        this.reflection = source["reflection"];
	        this.sort_order = source["sort_order"];
	        this.created_at = source["created_at"];
	    }
	}
	export class Interview {
	    id: number;
	    application_id: number;
	    round_name: string;
	    scheduled_at: string;
	    questions_and_notes: string;
	    outcome: string;
	    created_at: string;
	    qa_items?: QAItem[];
	
	    static createFrom(source: any = {}) {
	        return new Interview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.application_id = source["application_id"];
	        this.round_name = source["round_name"];
	        this.scheduled_at = source["scheduled_at"];
	        this.questions_and_notes = source["questions_and_notes"];
	        this.outcome = source["outcome"];
	        this.created_at = source["created_at"];
	        this.qa_items = this.convertValues(source["qa_items"], QAItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Application {
	    id: number;
	    company: string;
	    position: string;
	    status: string;
	    resume_version: string;
	    jd_text: string;
	    applied_at: string;
	    location: string;
	    salary_range: string;
	    contact_info: string;
	    notes: string;
	    reached_interview: number;
	    failed_round: number;
	    interview_count: number;
	    created_at: string;
	    updated_at: string;
	    interviews?: Interview[];
	
	    static createFrom(source: any = {}) {
	        return new Application(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.company = source["company"];
	        this.position = source["position"];
	        this.status = source["status"];
	        this.resume_version = source["resume_version"];
	        this.jd_text = source["jd_text"];
	        this.applied_at = source["applied_at"];
	        this.location = source["location"];
	        this.salary_range = source["salary_range"];
	        this.contact_info = source["contact_info"];
	        this.notes = source["notes"];
	        this.reached_interview = source["reached_interview"];
	        this.failed_round = source["failed_round"];
	        this.interview_count = source["interview_count"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	        this.interviews = this.convertValues(source["interviews"], Interview);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ApplicationInput {
	    company: string;
	    position: string;
	    resume_version: string;
	    jd_text: string;
	    applied_at: string;
	    location: string;
	    salary_range: string;
	    contact_info: string;
	    notes: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new ApplicationInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.company = source["company"];
	        this.position = source["position"];
	        this.resume_version = source["resume_version"];
	        this.jd_text = source["jd_text"];
	        this.applied_at = source["applied_at"];
	        this.location = source["location"];
	        this.salary_range = source["salary_range"];
	        this.contact_info = source["contact_info"];
	        this.notes = source["notes"];
	        this.status = source["status"];
	    }
	}
	
	export class InterviewInput {
	    application_id: number;
	    round_name: string;
	    scheduled_at: string;
	    questions_and_notes: string;
	    outcome: string;
	
	    static createFrom(source: any = {}) {
	        return new InterviewInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.application_id = source["application_id"];
	        this.round_name = source["round_name"];
	        this.scheduled_at = source["scheduled_at"];
	        this.questions_and_notes = source["questions_and_notes"];
	        this.outcome = source["outcome"];
	    }
	}
	
	export class QAItemInput {
	    interview_id: number;
	    question: string;
	    answer: string;
	    reflection: string;
	    sort_order: number;
	
	    static createFrom(source: any = {}) {
	        return new QAItemInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.interview_id = source["interview_id"];
	        this.question = source["question"];
	        this.answer = source["answer"];
	        this.reflection = source["reflection"];
	        this.sort_order = source["sort_order"];
	    }
	}
	export class ResumeVersionStat {
	    resume_version: string;
	    total_applied: number;
	    reached_interview: number;
	    interview_rate: number;
	
	    static createFrom(source: any = {}) {
	        return new ResumeVersionStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.resume_version = source["resume_version"];
	        this.total_applied = source["total_applied"];
	        this.reached_interview = source["reached_interview"];
	        this.interview_rate = source["interview_rate"];
	    }
	}
	export class Stats {
	    total_applications: number;
	    interviewing: number;
	    reached_interview: number;
	    offered: number;
	    resume_rejected: number;
	    interview_failed: number;
	    declined: number;
	    interview_rate: number;
	    offer_rate: number;
	    by_resume_version: ResumeVersionStat[];
	
	    static createFrom(source: any = {}) {
	        return new Stats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_applications = source["total_applications"];
	        this.interviewing = source["interviewing"];
	        this.reached_interview = source["reached_interview"];
	        this.offered = source["offered"];
	        this.resume_rejected = source["resume_rejected"];
	        this.interview_failed = source["interview_failed"];
	        this.declined = source["declined"];
	        this.interview_rate = source["interview_rate"];
	        this.offer_rate = source["offer_rate"];
	        this.by_resume_version = this.convertValues(source["by_resume_version"], ResumeVersionStat);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

