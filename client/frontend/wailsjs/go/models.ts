export namespace api {
	
	export class PreKey {
	    id: number;
	    publicKey: number[];
	
	    static createFrom(source: any = {}) {
	        return new PreKey(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.publicKey = source["publicKey"];
	    }
	}
	export class SignedPreKey {
	    id: number;
	    publicKey: number[];
	    signature: number[];
	
	    static createFrom(source: any = {}) {
	        return new SignedPreKey(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.publicKey = source["publicKey"];
	        this.signature = source["signature"];
	    }
	}
	export class KeyBundle {
	    registrationId: number;
	    identityKey: number[];
	    signedPreKey: SignedPreKey;
	    preKeys: PreKey[];
	
	    static createFrom(source: any = {}) {
	        return new KeyBundle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.registrationId = source["registrationId"];
	        this.identityKey = source["identityKey"];
	        this.signedPreKey = this.convertValues(source["signedPreKey"], SignedPreKey);
	        this.preKeys = this.convertValues(source["preKeys"], PreKey);
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

export namespace encryption {
	
	export class Envelope {
	    KeyID: number;
	    Version: number;
	    Iteration: number;
	    Signature: number[];
	
	    static createFrom(source: any = {}) {
	        return new Envelope(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.KeyID = source["KeyID"];
	        this.Version = source["Version"];
	        this.Iteration = source["Iteration"];
	        this.Signature = source["Signature"];
	    }
	}
	export class DecryptedMessage {
	    Plaintext: number[];
	    Ciphertext: number[];
	    Envelope?: Envelope;
	
	    static createFrom(source: any = {}) {
	        return new DecryptedMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Plaintext = source["Plaintext"];
	        this.Ciphertext = source["Ciphertext"];
	        this.Envelope = this.convertValues(source["Envelope"], Envelope);
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
	export class EncryptedMessage {
	    Serialized: number[];
	    Ciphertext: number[];
	    Envelope?: Envelope;
	
	    static createFrom(source: any = {}) {
	        return new EncryptedMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Serialized = source["Serialized"];
	        this.Ciphertext = source["Ciphertext"];
	        this.Envelope = this.convertValues(source["Envelope"], Envelope);
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

export namespace models {
	
	export class Conversation {
	    ID: string;
	    LastMessagePreview: string;
	    LastMessageSenderID: string;
	    LastMessageTimestamp: number;
	    RecipientIDs: string[];
	
	    static createFrom(source: any = {}) {
	        return new Conversation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.LastMessagePreview = source["LastMessagePreview"];
	        this.LastMessageSenderID = source["LastMessageSenderID"];
	        this.LastMessageTimestamp = source["LastMessageTimestamp"];
	        this.RecipientIDs = source["RecipientIDs"];
	    }
	}
	export class Message {
	    ID: string;
	    Text: string;
	    SenderID: string;
	    Timestamp: number;
	    Ciphertext: number[];
	    Envelope?: encryption.Envelope;
	
	    static createFrom(source: any = {}) {
	        return new Message(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Text = source["Text"];
	        this.SenderID = source["SenderID"];
	        this.Timestamp = source["Timestamp"];
	        this.Ciphertext = source["Ciphertext"];
	        this.Envelope = this.convertValues(source["Envelope"], encryption.Envelope);
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
	export class User {
	    ID: string;
	    Username: string;
	
	    static createFrom(source: any = {}) {
	        return new User(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Username = source["Username"];
	    }
	}

}

