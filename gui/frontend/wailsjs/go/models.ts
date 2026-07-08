export namespace main {
	
	export class AgentInfo {
	    id: string;
	    display_name: string;
	
	    static createFrom(source: any = {}) {
	        return new AgentInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.display_name = source["display_name"];
	    }
	}
	export class DoctorCheck {
	    group: string;
	    name: string;
	    status: string;
	    details: string;
	
	    static createFrom(source: any = {}) {
	        return new DoctorCheck(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.group = source["group"];
	        this.name = source["name"];
	        this.status = source["status"];
	        this.details = source["details"];
	    }
	}
	export class DoctorResult {
	    checks: DoctorCheck[];
	    summary: string;
	
	    static createFrom(source: any = {}) {
	        return new DoctorResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.checks = this.convertValues(source["checks"], DoctorCheck);
	        this.summary = source["summary"];
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
	export class RepoItem {
	    type: string;
	    name: string;
	    source: string;
	    cache_path: string;
	    commit_at: string;
	    branch: string;
	
	    static createFrom(source: any = {}) {
	        return new RepoItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.name = source["name"];
	        this.source = source["source"];
	        this.cache_path = source["cache_path"];
	        this.commit_at = source["commit_at"];
	        this.branch = source["branch"];
	    }
	}
	export class SCCFileStat {
	    type: string;
	    file: string;
	    language: string;
	    code: number;
	    complexity: number;
	
	    static createFrom(source: any = {}) {
	        return new SCCFileStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.file = source["file"];
	        this.language = source["language"];
	        this.code = source["code"];
	        this.complexity = source["complexity"];
	    }
	}
	export class SCCLangStat {
	    name: string;
	    count: number;
	    code: number;
	
	    static createFrom(source: any = {}) {
	        return new SCCLangStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.count = source["count"];
	        this.code = source["code"];
	    }
	}
	export class SCCResult {
	    repo: string;
	    languages: SCCLangStat[];
	    topFiles: SCCFileStat[];
	
	    static createFrom(source: any = {}) {
	        return new SCCResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repo = source["repo"];
	        this.languages = this.convertValues(source["languages"], SCCLangStat);
	        this.topFiles = this.convertValues(source["topFiles"], SCCFileStat);
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

