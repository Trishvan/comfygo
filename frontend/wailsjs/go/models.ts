export namespace orchestrator {
	
	export class GenerationParams {
	    prompt: string;
	    negativePrompt: string;
	    modelPath: string;
	    vaePath: string;
	    steps: number;
	    cfgScale: number;
	    seed: number;
	    width: number;
	    height: number;
	    samplerName: string;
	    loraPaths: string[];
	    loraScales: number[];
	
	    static createFrom(source: any = {}) {
	        return new GenerationParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.prompt = source["prompt"];
	        this.negativePrompt = source["negativePrompt"];
	        this.modelPath = source["modelPath"];
	        this.vaePath = source["vaePath"];
	        this.steps = source["steps"];
	        this.cfgScale = source["cfgScale"];
	        this.seed = source["seed"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.samplerName = source["samplerName"];
	        this.loraPaths = source["loraPaths"];
	        this.loraScales = source["loraScales"];
	    }
	}
	export class HistoryEntry {
	    id: number;
	    params: GenerationParams;
	    status: string;
	    filename: string;
	    error: string;
	    width: number;
	    height: number;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new HistoryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.params = this.convertValues(source["params"], GenerationParams);
	        this.status = source["status"];
	        this.filename = source["filename"];
	        this.error = source["error"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.createdAt = source["createdAt"];
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
	export class QueueItem {
	    id: number;
	    params: GenerationParams;
	    status: string;
	    progress: number;
	    outputPath: string;
	    error: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new QueueItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.params = this.convertValues(source["params"], GenerationParams);
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.outputPath = source["outputPath"];
	        this.error = source["error"];
	        this.createdAt = source["createdAt"];
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
	export class SystemStats {
	    ramTotalGB: number;
	    ramUsedGB: number;
	    ramPercent: number;
	    vramTotalGB: number;
	    vramUsedGB: number;
	    vramPercent: number;
	
	    static createFrom(source: any = {}) {
	        return new SystemStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ramTotalGB = source["ramTotalGB"];
	        this.ramUsedGB = source["ramUsedGB"];
	        this.ramPercent = source["ramPercent"];
	        this.vramTotalGB = source["vramTotalGB"];
	        this.vramUsedGB = source["vramUsedGB"];
	        this.vramPercent = source["vramPercent"];
	    }
	}

}

