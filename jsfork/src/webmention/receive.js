
const got = require('got')
const config = require('./../config')
const fsp = require('fs').promises
const md5 = require('md5')
const { mf2 } = require("microformats-parser");
const dayjs = require('dayjs')
const utc = require('dayjs/plugin/utc')
dayjs.extend(utc)

const log = require('pino')()

function validate(request) {
	// DONE
}

async function isValidTargetUrl(target) {
	try {
		await got(target)
		return true
	} catch(unknownTarget) {
	}
	log.warn(` ABORT: invalid target url`)
	return false
}

function asPath(source, target) {
	const filename = md5(`source=${source},target=${target}`)
	const domain = config.allowedWebmentionSources.find(d => target.indexOf(d) >= 0)
	return `data/${domain}/${filename}.json`
}

async function deletePossibleOlderWebmention(source, target) {
	try {
		await fsp.unlink(asPath(source, target))
	} catch(e) {
		// does not matter, file not there. 
	}
}

async function saveWebmentionToDisk(source, target, mentiondata) {
	await fsp.writeFile(asPath(source, target), mentiondata, 'utf-8')
}

function publishedNow() {
	return dayjs.utc().utcOffset(config.utcOffset).format("YYYY-MM-DDTHH:mm:ss")
}

function parseBodyAsIndiewebSite(source, target, hEntry) {
	function shorten(txt) {
		if(!txt || txt.length <= 250) return txt
		return txt.substring(0, 250) + "..."
	}

	const name = hEntry.properties?.name?.[0]
	const authorPropName = hEntry.properties?.author?.[0]?.properties?.name?.[0]
	const authorValue = hEntry.properties?.author?.[0]?.value
	const picture = hEntry.properties?.author?.[0]?.properties?.photo?.[0]
	const summary = hEntry.properties?.summary?.[0]
	const contentEntry = hEntry.properties?.content?.[0]?.value
	const bridgyTwitterContent = hEntry.properties?.["bridgy-twitter-content"]?.[0]
	const publishedDate = hEntry.properties?.published?.[0]
	const uid = hEntry.properties?.uid?.[0]
	const url = hEntry.properties?.url?.[0]
	const type = hEntry.properties?.["like-of"]?.length ? "like" : (hEntry.properties?.["bookmark-of"]?.length ? "bookmark" : "mention" )

	return {
		author: {
			name: authorPropName ? authorPropName : authorValue,
			picture: picture?.value ? picture?.value : picture
		},
		name: name,
		content: bridgyTwitterContent ? shorten(bridgyTwitterContent) : (summary ? shorten(summary) : shorten(contentEntry)),
		published: publishedDate ? publishedDate : publishedNow(),
		type,
		// Mastodon uids start with "tag:server", but we do want indieweb uids from other sources 
		url: uid && uid.startsWith("http") ? uid : (url ? url : source),
		source,
		target
	}
}

function parseBodyAsNonIndiewebSite(source, target, body) {
	const title = body.match(/<title>(.*?)<\/title>/)?.splice(1, 1)[0]

	return {
		author: {
			name: source
		},
		name: title,
		content: title,
		published: publishedNow(),
		url: source,
		type: "mention",
		source,
		target
	}
}

async function processSourceBody(body, source, target) {
	if(body.indexOf(target) === -1) {
		log.warn(` ABORT: no mention of ${target} found in html src of source`)
		return
	}

	// fiddle: https://aimee-gm.github.io/microformats-parser/
	const microformat = mf2(body, {
		// WHY? crashes on relative URL, should be injected using Jest. Don't care. 
		baseUrl: source.startsWith("http") ? source : `http://localhost/${source}`
	})
	const hEntry = microformat.items.filter(itm => itm?.type?.includes("h-entry"))?.[0]

	const data = hEntry ? parseBodyAsIndiewebSite(source, target, hEntry) : parseBodyAsNonIndiewebSite(source, target, body)
	await saveWebmentionToDisk(source, target, JSON.stringify(data))
	log.info(` OK: webmention processed`)
}

async function receive(body) {
	if(!isValidTargetUrl(body.target)) return

	let src = { body: "" }
	try {
		src = await got(body.source)
	} catch(unknownSource) {
		log.warn(` ABORT: invalid source url: ` + unknownSource)
		await deletePossibleOlderWebmention(body.source, body.target)
		return
	}
	await processSourceBody(src.body, body.source, body.target)
} 

module.exports = {
	receive,
	validate
}