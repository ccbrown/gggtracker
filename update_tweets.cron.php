<?php
define('IN_GGGT', true);
require_once "./includes/common.inc.php";

if (!isset($_GET['secret']) || $_GET['secret'] != $_['cron']['secret']) {
	die();
}

$result = $SQL->query("SELECT twitter_tweet_id FROM tweets ORDER BY twitter_tweet_id DESC LIMIT 0, 1");

$last_tweet_id = ($SQL->num_rows($result) > 0 ? $SQL->result($result, 0) : 0);

$oauth = new OAuth($_['twitter']['consumer_key'], $_['twitter']['consumer_secret']);
$oauth->setToken($_['twitter']['access_token'], $_['twitter']['access_secret']);

$success = false;

try {
	$success = $oauth->fetch('https://api.twitter.com/1.1/statuses/user_timeline.json?screen_name='.urlencode($_['poe']['twitter_name']).($last_tweet_id ? '&since_id='.urlencode($last_tweet_id) : ''));
} catch (OAuthException $e) {
	echo 'Exception: '.$e->getMessage().'<br />';
}

if (!$success) {
	die('Failure');
}

$data = json_decode($oauth->getLastResponse());

if ($data === null) {
	die('Failure');
}

// twitter tries to escape stuff for us >:(
function clean_tweet($tweet) {
	$tweet = str_replace('&lt;', '<', $tweet);
	$tweet = str_replace('&gt;', '>', $tweet);
	$tweet = str_replace('&amp;', '&', $tweet);

	return $tweet;
}

foreach ($data as $tweet) {
	$fields = array(
		'twitter_tweet_id' => $tweet->id,
		'tweet_time'       => strtotime($tweet->created_at),
		'tweet_text'       => clean_tweet($tweet->text),
	);

	$SQL->query("INSERT INTO tweets SET ".$SQL->compile_set_fields($fields));
}

echo 'Added '.count($data).' tweet'.(count($data) == 1 ? '' : 's').'!';

?>
