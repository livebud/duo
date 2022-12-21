const parser = require('./jsparser')
const assert = require('assert')
const path = require('path')
const fs = require('fs')

const samplesDir = path.join(__dirname, `../submodule/svelte/test/parser/samples`)

describe("svelte/test/parser/samples", function () {
  const testdirs = fs.readdirSync(samplesDir)
  testdirs.forEach(runTest)
})

function tryRead(path) {
  try {
    return fs.readFileSync(path, 'utf8')
  } catch (err) {
    return ""
  }
}

function deleteKeys(obj, keys) {
  for (let prop in obj) {
    for (let key of keys) {
      if (prop === key) {
        delete obj[key]
      }
    }
    if (typeof obj[prop] === 'object') {
      obj[prop] = deleteKeys(obj[prop], keys)
    }
    if (Array.isArray(obj[prop])) {
      for (let i = 0; i < obj[prop].length; i++) {
        if (typeof obj[prop][i] === 'object') {
          obj[prop][i] = deleteKeys(obj[prop][i], keys)
        }
      }
    }
  }
  return obj
}

function runTest(dir) {
  it(dir, function () {
    const input = fs.readFileSync(path.join(samplesDir, dir, "input.svelte"), 'utf8')
    const output = tryRead(path.join(samplesDir, dir, "output.json"))
    const error = tryRead(path.join(samplesDir, dir, "error.json"))
    const expect = JSON.parse(output || error)
    deleteKeys(expect, ['start', 'end', 'loc'])
    const actual = {
      html: parser.parse(input.trim())
    }
    assert.deepEqual(actual, expect)
  })
}

