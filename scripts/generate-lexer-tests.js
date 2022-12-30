const path = require('path')
const fs = require('fs')

const samplesDir = path.join(__dirname, `../submodule/svelte/test/parser/samples`)
const testdirs = fs.readdirSync(samplesDir)

for (let dir of testdirs) {
  const input = fs.readFileSync(path.join(samplesDir, dir, "input.svelte"), 'utf8')
  console.log(`is.Equal(lex(${JSON.stringify(input)}), ` + "``" + `) // ${dir} ${JSON.stringify(input).slice(1, -1)}`)
}
