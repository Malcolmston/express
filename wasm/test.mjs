// Node smoke test: builds must be run first (see build.sh). Verifies the Go
// utility implementations are reachable from JS through wasm.
import assert from 'node:assert';
import { loadExpress } from './express.mjs';

const ex = await loadExpress();

// ms
assert.strictEqual(ex.ms('2h'), 7200000, 'ms("2h") === 7200000');
assert.strictEqual(ex.ms('1d'), 86400000, 'ms("1d") === 86400000');
assert.strictEqual(ex.ms('not a duration'), null, 'ms(invalid) === null');
assert.strictEqual(ex.msFormat(7200000), '2h', 'msFormat(7200000) === "2h"');

// bytes (binary units: 1KB = 1024)
assert.strictEqual(ex.bytes('1KB'), 1024, 'bytes("1KB") === 1024');
assert.strictEqual(ex.bytesFormat(1536), '1.5KB', 'bytesFormat(1536) === "1.5KB"');
assert.strictEqual(ex.bytes('nope'), null, 'bytes(invalid) === null');

// uuid
const id = ex.uuidV4();
assert.match(id, /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/,
  'uuidV4 is a valid v4 uuid');
assert.strictEqual(ex.uuidValidate(id), true, 'uuidValidate(generated) === true');
assert.strictEqual(ex.uuidValidate('nope'), false, 'uuidValidate(garbage) === false');

// nanoid
const nid = ex.nanoid();
assert.strictEqual(nid.length, 21, 'nanoid default length is 21');
assert.match(nid, /^[A-Za-z0-9_-]{21}$/, 'nanoid uses the url-safe alphabet');

// slugify
assert.strictEqual(ex.slugify('Hello World'), 'hello-world', 'slugify lowercases by default');
assert.strictEqual(ex.slugify('Hello World', false), 'Hello-World', 'slugify(_, false) keeps case');

// validatorjs
assert.strictEqual(ex.isEmail('a@b.com'), true, 'isEmail("a@b.com") === true');
assert.strictEqual(ex.isEmail('nope'), false, 'isEmail("nope") === false');
assert.strictEqual(ex.isURL('https://example.com'), true, 'isURL(valid) === true');

// pluralize
assert.strictEqual(ex.plural('house'), 'houses', 'plural("house") === "houses"');

// titlecase
assert.strictEqual(ex.titleCase('hello world'), 'Hello World', 'titleCase capitalizes words');

console.log('express wasm adapter: all JS-side assertions passed');
console.log(`  sample: uuid=${id} nanoid=${nid} slug=${ex.slugify('Express in JS via WASM')}`);
process.exit(0);
