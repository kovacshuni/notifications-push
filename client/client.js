"use strict";

const EventSource = require('eventsource');

let eventSource;

function error(type, d){
	let logStr = `event=PUSH_EVENT_ERROR error=${type}`;
	console.log(logStr, d);
}

function onMessage(e){
	console.log(`event=PUSH_EVENT_RECEIVED`, e.data);
	let apiUrl;
	let data = JSON.parse(e.data);

	if(data.length === 0){
		console.log('event=EMPTY_DATA');
	}
	//do stuff
}

function start(){
	if(!process.env.PUSH_API_URL){
		throw new Error('PUSH_API_URL env var missing!');
	}
	if(!process.env.COCO_API_AUTHORIZATION){
		throw new Error('COCO_API_AUTHORIZATION env var missing');
	}


	console.log(`event=PUSH_API_CONNECT url=${process.env.PUSH_API_URL}`);

	eventSource = new EventSource(
		process.env.PUSH_API_URL,
		{
			headers:{
				'Authorization' : process.env.COCO_API_AUTHORIZATION
			}
		}
	);
	eventSource.onmessage = onMessage;
	eventSource.onopen = () => console.log(`event=PUSH_API_CONNECTION_OPEN`);
	eventSource.onerror = err => {
		let stack = typeof err.stack === 'string' ? err.stack.replace(/\n/g, '; ') : '';
		let status = err.status || null;
		error('EVENT_SOURCE_ERROR', {message:err.message, stack:stack, status:status});
	};
}

function stop(){
	eventSource.close();
}
start()
