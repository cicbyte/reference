export namespace main {
	
	export class AgentInfo {
	    id: string;
	    displayName: string;
	    baseDir: string;
	    fileCount: number;
	
	    static createFrom(source: any = {}) {
	        return new AgentInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.displayName = source["displayName"];
	        this.baseDir = source["baseDir"];
	        this.fileCount = source["fileCount"];
	    }
	}
	export class AppConfigDTO {
	    reposPath: string;
	    wikiPath: string;
	    // Go type: struct { Proxy string "json:\"proxy\""; GitProxy string "json:\"gitProxy\""; Timeout int "json:\"timeout\"" }
	    network: any;
	    // Go type: struct { Level string "json:\"level\""; MaxSize int "json:\"maxSize\""; MaxBackups int "json:\"maxBackups\""; MaxAge int "json:\"maxAge\""; Compress bool "json:\"compress\"" }
	    log: any;
	    // Go type: struct { Config string "json:\"config\""; Db string "json:\"db\""; LogDir string "json:\"logDir\""; Repos string "json:\"repos\""; Wiki string "json:\"wiki\"" }
	    paths: any;
	
	    static createFrom(source: any = {}) {
	        return new AppConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reposPath = source["reposPath"];
	        this.wikiPath = source["wikiPath"];
	        this.network = this.convertValues(source["network"], Object);
	        this.log = this.convertValues(source["log"], Object);
	        this.paths = this.convertValues(source["paths"], Object);
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
	export class BrowserFileNode {
	    name: string;
	    path: string;
	    isDir: boolean;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new BrowserFileNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.isDir = source["isDir"];
	        this.size = source["size"];
	    }
	}
	export class BrowserFileResult {
	    content: string;
	    lines: number;
	    binary: boolean;
	    notFound: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BrowserFileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.content = source["content"];
	        this.lines = source["lines"];
	        this.binary = source["binary"];
	        this.notFound = source["notFound"];
	    }
	}
	export class CacheTopItem {
	    name: string;
	    path: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new CacheTopItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	    }
	}
	export class CachedRepoItem {
	    name: string;
	    cachePath: string;
	    type: string;
	    host: string;
	    namespace: string;
	    repoName: string;
	    exists: boolean;
	    size: number;
	    refCount: number;
	    projects: string[];
	    branch: string;
	    commit: string;
	
	    static createFrom(source: any = {}) {
	        return new CachedRepoItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.cachePath = source["cachePath"];
	        this.type = source["type"];
	        this.host = source["host"];
	        this.namespace = source["namespace"];
	        this.repoName = source["repoName"];
	        this.exists = source["exists"];
	        this.size = source["size"];
	        this.refCount = source["refCount"];
	        this.projects = source["projects"];
	        this.branch = source["branch"];
	        this.commit = source["commit"];
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
	export class ProjectInfo {
	    dir: string;
	    name: string;
	    exists: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProjectInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dir = source["dir"];
	        this.name = source["name"];
	        this.exists = source["exists"];
	    }
	}
	export class ProjectItem {
	    dir: string;
	    name: string;
	    exists: boolean;
	    initialized: boolean;
	    agents: string[];
	    repoCount: number;
	    brokenCount: number;
	
	    static createFrom(source: any = {}) {
	        return new ProjectItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dir = source["dir"];
	        this.name = source["name"];
	        this.exists = source["exists"];
	        this.initialized = source["initialized"];
	        this.agents = source["agents"];
	        this.repoCount = source["repoCount"];
	        this.brokenCount = source["brokenCount"];
	    }
	}
	export class RepoDiagnosis {
	    refName: string;
	    linkName: string;
	    type: string;
	    remoteUrl: string;
	    cachePath: string;
	    localPath: string;
	    branch: string;
	    targetExists: boolean;
	    linkExists: boolean;
	    wikiExists: boolean;
	    status: string;
	    suggestion: string;
	
	    static createFrom(source: any = {}) {
	        return new RepoDiagnosis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.refName = source["refName"];
	        this.linkName = source["linkName"];
	        this.type = source["type"];
	        this.remoteUrl = source["remoteUrl"];
	        this.cachePath = source["cachePath"];
	        this.localPath = source["localPath"];
	        this.branch = source["branch"];
	        this.targetExists = source["targetExists"];
	        this.linkExists = source["linkExists"];
	        this.wikiExists = source["wikiExists"];
	        this.status = source["status"];
	        this.suggestion = source["suggestion"];
	    }
	}
	export class RepoItem {
	    type: string;
	    name: string;
	    source: string;
	    cache_path: string;
	    commit_at: string;
	    branch: string;
	    remoteUrl: string;
	    cacheExists: boolean;
	
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
	        this.remoteUrl = source["remoteUrl"];
	        this.cacheExists = source["cacheExists"];
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
	export class WikiEntry {
	    repoName: string;
	    platform: string;
	    namespace: string;
	    source: string;
	    relPath: string;
	    fullPath: string;
	    fileName: string;
	    commit: string;
	    branch: string;
	    description: string;
	    status: string;
	    exploredAt: string;
	    modifiedAt: string;
	    gitStatus: string;
	
	    static createFrom(source: any = {}) {
	        return new WikiEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repoName = source["repoName"];
	        this.platform = source["platform"];
	        this.namespace = source["namespace"];
	        this.source = source["source"];
	        this.relPath = source["relPath"];
	        this.fullPath = source["fullPath"];
	        this.fileName = source["fileName"];
	        this.commit = source["commit"];
	        this.branch = source["branch"];
	        this.description = source["description"];
	        this.status = source["status"];
	        this.exploredAt = source["exploredAt"];
	        this.modifiedAt = source["modifiedAt"];
	        this.gitStatus = source["gitStatus"];
	    }
	}

}

