export namespace models {
	
	export class Conversation {
	    ID: string;
	    LastMessagePreview: string;
	    LastMessageSenderID: string;
	    LastMessageTimestamp: number;
	    ParticipantIDs: string[];
	
	    static createFrom(source: any = {}) {
	        return new Conversation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.LastMessagePreview = source["LastMessagePreview"];
	        this.LastMessageSenderID = source["LastMessageSenderID"];
	        this.LastMessageTimestamp = source["LastMessageTimestamp"];
	        this.ParticipantIDs = source["ParticipantIDs"];
	    }
	}
	export class Message {
	    ID: string;
	    Text: string;
	    SenderID: string;
	    Timestamp: number;
	
	    static createFrom(source: any = {}) {
	        return new Message(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Text = source["Text"];
	        this.SenderID = source["SenderID"];
	        this.Timestamp = source["Timestamp"];
	    }
	}
	export class User {
	    ID: string;
	    Username: string;
	    Avatar: string;
	
	    static createFrom(source: any = {}) {
	        return new User(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Username = source["Username"];
	        this.Avatar = source["Avatar"];
	    }
	}

}

