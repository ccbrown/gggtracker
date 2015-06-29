<?php
if (!defined('IN_GGGT')) { die(); }

function fetch_page($url, $cookie = '', $redirect_count = 0) {
	$ch = curl_init();
	curl_setopt_array($ch, array(
		CURLOPT_URL => $url,
		CURLOPT_HEADER => 0,
		CURLOPT_RETURNTRANSFER => true,
		CURLOPT_TIMEOUT => 20,
		CURLOPT_COOKIE => $cookie,
	));

	$result = curl_exec($ch);

	$info = curl_getinfo($ch);
	
	curl_close($ch);

	if ($info['http_code'] == 301 || $info['http_code'] == 302) {
		return $redirect_count < 5 ? fetch_page($info['redirect_url'], $cookie, $redirect_count + 1) : null;
	}

	return $result;
}

function inner_html($element) { 
	$innerHTML = ""; 
	$children = $element->childNodes; 
	foreach ($children as $child) { 
		$tmp_dom = new DOMDocument(); 
		$tmp_dom->appendChild($tmp_dom->importNode($child, true)); 
		$innerHTML .= trim($tmp_dom->saveHTML()); 
	} 
	return $innerHTML;
} 

function get_posts($poster, $page) {
	global $_, $SQL;

	$html = fetch_page('http://www.pathofexile.com/account/view-posts/'.urlencode($poster).'/page/'.urlencode($page), 'PHPSESSID='.$_['poe']['forum_sessid']);
	
	if (!$html) {
		return false;
	}
	
	$dom = new DOMDocument;
	$dom->loadHTML($html);
	
	$finder = new DomXPath($dom);
	
	$post_list_node = $finder->query("//*[contains(concat(' ', normalize-space(@class), ' '), ' forumPostListTable ')]")->item(0);
	
	if (!$post_list_node) {
		return false;
	}

	$post_nodes = $finder->query(".//tr", $post_list_node);
	
	$posts = array();
	
	for ($i = 0; $i < $post_nodes->length; ++$i) {
		$post_node = $post_nodes->item($i);

		$content_node = $finder->query(".//*[contains(concat(' ', normalize-space(@class), ' '), ' content ')]", $post_node)->item(0);
		if (!$content_node) {
			return false;
		}

		$date_node = $finder->query(".//*[contains(concat(' ', normalize-space(@class), ' '), ' post_date ')]", $post_node)->item(0);
		if (!$date_node) {
			return false;
		}
		
		$body = inner_html($content_node);
		$time = inner_html($date_node);

		$post_info_node = $finder->query(".//*[contains(concat(' ', normalize-space(@class), ' '), ' post_info ')]", $post_node)->item(0);
		if (!$post_info_node) {
			return false;
		}
		
		$link_nodes = $finder->query(".//a", $post_info_node);

		$post_id     = 0;
		$thread_id   = 0;
		$forum_id    = 0;
		$page_number = 0;

		$thread_name = "";
		$forum_name  = "";

		for ($j = 0; $j < $link_nodes->length; ++$j) {
			$link_node = $link_nodes->item($j);

			if (!$link_node->attributes) {
				continue;
			}
			
			$href = $link_node->attributes->getNamedItem('href');
			
			if (!$href) {
				continue;
			}

			if (preg_match('/^\/forum\/view-thread\/([0-9]+)\/page\/([0-9]+)#p([0-9]+)/', $href->nodeValue, $matches)) {
				$thread_id   = $matches[1];
				$page_number = $matches[2];
				$post_id     = $matches[3];
			} else if (preg_match('/^\/forum\/view-thread\/[0-9]+/', $href->nodeValue, $matches)) {
				$thread_name = $link_node->nodeValue;
			} else if (preg_match('/^\/forum\/view-forum\/([0-9]+)/', $href->nodeValue, $matches)) {
				$forum_id = $matches[1];
				$forum_name = $link_node->nodeValue;
			}
		}
		
		if (!post_id || !thread_id || !forum_id) {
			return false;
		}
		
		$posts[] = array(
			'post_body'         => $body,
			'post_time'         => strtotime($time.' '.$_['poe']['forum_timezone']),
			'post_author'       => $poster,
			'forum_post_id'     => $post_id,
			'forum_thread_id'   => $thread_id,
			'thread_name'       => $thread_name,
			'forum_forum_id'    => $forum_id,
			'forum_name'        => $forum_name,
			'forum_page_number' => $page_number,
		);
	}
	
	return $posts;
}

function index_poster($poster) {	
	global $_, $SQL;

	$page = 1;
	
	while (($posts = get_posts($poster, $page)) !== false && count($posts) > 0) {
		foreach ($posts as $post) {
			$result = $SQL->query("SELECT post_id FROM posts WHERE forum_post_id = '".$SQL->escape_string($post['forum_post_id'])."'");
			if ($SQL->num_rows($result) > 0) {
				return true;
			}
			
			$SQL->query("INSERT INTO posts SET ".$SQL->compile_set_fields($post));
		}
		
		++$page;
	}
	
	return ($posts !== false);
}
?>