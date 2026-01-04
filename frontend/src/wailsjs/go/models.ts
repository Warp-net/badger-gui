export namespace main {
	
	export class AppMessage {
	    type: string;
	    body: string;
	
	    static createFrom(source: any = {}) {
	        return new AppMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.body = source["body"];
	    }
	}

}

