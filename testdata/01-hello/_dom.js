// Generated by `node scripts/generate-dom.js`. Do not edit.

import * as $ from "svelte/internal/client";

var root = $.template(`<h1>hello world</h1>`);

export default function Input($$anchor) {
	var h1 = root();

	$.append($$anchor, h1);
}