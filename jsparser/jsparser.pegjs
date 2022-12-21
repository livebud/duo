// Simple Arithmetics Grammar
// ==========================
//
// Accepts expressions like "2 * (3 + 4)" and computes their value.

Fragment = children:TemplateNode* {
  return {
    type: "Fragment",
    children: children,
  }
}

TemplateNode =
  // ConstTag
  // / DebugTag
  MustacheTag /
  // / BaseNode
  Element /
  // / SpreadAttribute
  // / Directive
  // / Transition
  // / Comment
  Text

Text = data:[^{<]+ {
  return {
    type: "Text",
    data: data.join(''),
    raw: data.join(''),
  }
}

MustacheTag = '{' expr:JSExpr '}' {
  return {
    type:"MustacheTag",
    expression: expr,
  }
}

Attribute = name:AttributeName _ '=' value:AttributeValue* {
  return {
    type: "Attribute",
    name: name,
    value: value,
  }
}

AttributeName = [a-z]+ {
  return text()
}

AttributeValue = AttributeText

AttributeText = ['"] text:[^'"]* ['"] {
  return {
    data: text.join(''),
    raw: text.join(''),
    type: "Text",
  }
}

Element = '<' _ tag:TagName list:(ws Attribute)* _ selfclosing:"/"? ">" children:TemplateNode* '</' TagName '>' {
  return {
    type: tag.type,
    name: tag.name,
    attributes: list.map(item => item[1]),
    children: children,
  }
}

TagName =
  InlineComponent /
  // SlotTemplate /
  // Title /
  // Head /
  // Options /
  // Window /
  // Body /
  // Slot /
  HtmlElement

InlineComponent = name:[A-Z][A-Za-z0-9_$]* {
  return {
    type: "InlineComponent",
    name: name,
  }
}

HtmlElement = name:[a-z][A-Za-z0-9_$]* {
  return {
    type: "Element",
    name: text(),
  }
}

JSExpr = JSIdentifier

JSIdentifier = [a-zA-Z$_][a-zA-Z0-9$_]* {
  return {
    type: "Identifier",
    name: text()
  }
}

ws = [ \t\n\r]+

_ "whitespace"
  = [ \t\n\r]*
