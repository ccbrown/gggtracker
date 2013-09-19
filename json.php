<?php
define('IN_GGGT', true);
require_once "./includes/common.inc.php";

header('Content-type: application/json');

$result = $SQL->query("
	SELECT p.*, b.post_count 
	FROM posts p 
	JOIN (
			SELECT MAX(forum_post_id) max_forum_post_id, COUNT(post_id) post_count 
			FROM posts 
			GROUP BY forum_thread_id 
			ORDER BY max_forum_post_id DESC 
			LIMIT 0, 50
		) b 
	ON b.max_forum_post_id = p.forum_post_id
");

$data['posts'] = array();

while ($row = $SQL->fetch_assoc($result)) {
	$data['posts'][] = array(
		'author'      => $row['post_author'],
		'time'        => $row['post_time'], 
		'forum_name'  => $row['forum_name'],
		'forum_id'    => $row['forum_forum_id'], 
		'thread_name' => $row['thread_name'],
		'thread_id'   => $row['forum_thread_id'],
		'post_id'     => $row['forum_post_id'], 
		'post_count'  => $row['post_count'], 
		'url'         => 'http://www.pathofexile.com/forum/view-thread/'.urlencode($row['forum_thread_id']).'/page/'.urlencode($row['forum_page_number']).'#p'.urlencode($row['forum_post_id']),
	);
}

die(json_encode($data));
?>