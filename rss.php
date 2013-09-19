<?php
define('IN_GGGT', true);
require_once "./includes/common.inc.php";

header('Content-type: application/rss+xml');

$xml = new XMLWriter;
$xml->openURI('php://output');
$xml->startDocument('1.0', 'UTF-8');

$xml->startElement('rss');
$xml->writeAttribute('version', '2.0');
$xml->writeAttribute('xml:base', 'http://www.pathofexile.com');
$xml->startElement('channel');

$xml->writeElement('title', 'GGG Tracker Forum Feed');
$xml->writeElement('description', 'Latest Forum Posts by Grinding Gear Games');
$xml->writeElement('link', 'http://www.gggtracker.com');

$result = $SQL->query("SELECT * FROM posts WHERE post_time >= '".(time() - 60 * 60 * 24 * 4)."' ORDER BY post_time DESC, forum_post_id DESC");

while ($row = $SQL->fetch_assoc($result)) {
	$xml->startElement('item');

	$xml->writeElement('title', $row['post_author'].' - '.$row['thread_name']);
	$xml->writeElement('link', 'http://www.pathofexile.com/forum/view-thread/'.urlencode($row['forum_thread_id']).'/page/'.urlencode($row['forum_page_number']).'#p'.urlencode($row['forum_post_id']));
	$xml->writeElement('guid', 'poe-forum-post-'.urlencode($row['post_id']).'-'.urlencode($row['forum_post_id']));
	$xml->writeElement('description', $row['post_body']);
	$xml->writeElement('pubDate', date('r', $row['post_time']));

	$xml->endElement(); // item
}

$xml->endElement(); // channel
$xml->endElement(); // rss

$xml->endDocument();
?>