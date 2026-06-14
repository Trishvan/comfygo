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
	    }
	}

}

