// Generated by `node scripts/generate-dom.js`. Do not edit.

import * as $ from "svelte/internal/client";

function addToList(_, todoList, newItem) {
	todoList($.proxy([
		...todoList(),
		{ text: $.get(newItem), status: false }
	]));

	$.set(newItem, "");
}

var root_1 = $.template(`<input type="checkbox"> <span class="svelte-1mf3hor"> </span> <span>❌</span> <br>`, 1);
var root = $.template(`<input type="text" placeholder="new todo item"> <button>Add</button> <br> <!>`, 1);

export default function Input($$anchor, $$props) {
	$.push($$props, true);

	let todoList = $.prop($$props, "todoList", 15, () => $.proxy([]));
	let newItem = $.source("");

	function removeFromList(index) {
		todoList().splice(index, 1);
	}

	var fragment = root();
	var input = $.first_child(fragment);

	$.remove_input_defaults(input);

	var button = $.sibling($.sibling(input, true));

	button.__click = [addToList, todoList, newItem];

	var br = $.sibling($.sibling(button, true));
	var node = $.sibling($.sibling(br, true));

	$.each(node, 65, todoList, $.index, ($$anchor, item, index) => {
		var fragment_1 = root_1();
		var input_1 = $.first_child(fragment_1);

		$.remove_input_defaults(input_1);

		var span = $.sibling($.sibling(input_1, true));
		var text = $.child(span);

		$.reset(span);

		var span_1 = $.sibling($.sibling(span, true));

		span_1.__click = () => removeFromList(index);

		var br_1 = $.sibling($.sibling(span_1, true));

		$.template_effect(() => {
			$.toggle_class(span, "checked", $.unwrap(item).status);
			$.set_text(text, $.unwrap(item).text);
		});

		$.bind_checked(input_1, () => $.unwrap(item).status, ($$value) => ($.unwrap(item).status = $$value));
		$.append($$anchor, fragment_1);
	});

	$.template_effect(() => button.disabled = $.get(newItem) === "");
	$.bind_value(input, () => $.get(newItem), ($$value) => $.set(newItem, $$value));
	$.append($$anchor, fragment);
	$.pop();
}

$.delegate(["click"]);