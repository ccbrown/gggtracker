<?php
define('IN_GGGT', true);
require_once "./includes/common.inc.php";
require_once "./includes/header.inc.php";
?>

<?php
$result = $SQL->query("SELECT COUNT(DISTINCT forum_thread_id) FROM posts");
$thread_count = $SQL->result($result, 0);
$threads_per_page = 50;
$pages = ceil((float)$thread_count / $threads_per_page);
$page = !isset($_GET['page']) ? 1 :
		($_GET['page'] == 'first' ? 1 :
		($_GET['page'] == 'last' ? $pages :
		(is_numeric($_GET['page']) && intval($_GET['page']) > 0 && intval($_GET['page']) <= $pages ? intval($_GET['page']) :
		1)));

$page_links = '<a href="?page=first">First</a> ';
if ($page - 10 >= 1) {
	$page_links .= '... ';
}
for ($i = $page - 9; $i < $page + 10 && $i <= $pages; ++$i) {
	if ($i < 1) {
		$i = 1;
	}
	$page_links .= ($i == $page ? '<b>'.$i.'</b> ' : '<a href="?page='.$i.'">'.$i.'</a> ');
}
if ($page + 10 <= $pages) {
	$page_links .= '... ';
}
$page_links .= '<a href="?page=last">Last</a> ';
?>

<center>
	<div class="container">
		<div class="news-box">
			<h1>News</h1>
			<table class="list headless">	
				<?php
				$result = $SQL->query("
					SELECT p.*
					FROM posts p 
					JOIN (
							SELECT MIN(forum_post_id) min_forum_post_id
							FROM posts 
							WHERE forum_forum_id = '".$SQL->escape_string($_['poe']['news_forum_id'])."'
							GROUP BY forum_thread_id 
							ORDER BY min_forum_post_id DESC 
							LIMIT 0, 5
						) b 
					ON b.min_forum_post_id = p.forum_post_id
				");
				
				$alt = false;
				while ($row = $SQL->fetch_assoc($result)) {
					?>
					<tr<?= $alt ? ' class="alt"' : '' ?>>
						<td>
							<a href="http://www.pathofexile.com/forum/view-thread/<?= htmlspecialchars($row['forum_thread_id']) ?>"><?= htmlspecialchars($row['thread_name']) ?></a> 
							<small class="subtle js-time" data-time="<?= htmlspecialchars($row['post_time']) ?>"><?= create_date($row['post_time']) ?></small>
						</td>
					</tr>
					<?php
					$alt = !$alt;
				}
				?>
			</table>
			<div class="right"><small><a href="http://www.pathofexile.com/news/archive">News Archive</a></small></div>
		</div>
	
		<div class="tweets-box">
			<h1>Twitter</h1>
			<table class="list headless">	
				<?php
				$result = $SQL->query("
					SELECT t.*
					FROM tweets t
					ORDER BY twitter_tweet_id DESC
					LIMIT 0, 5
				");
				
				$alt = false;
				while ($row = $SQL->fetch_assoc($result)) {
					$tweet = htmlspecialchars($row['tweet_text']);
					$tweet = preg_replace('/(^|[^a-zA-Z0-9\p{L}])([a-zA-Z]+\:\/\/[a-zA-Z0-9\/\\%\-\.]+)/im', '$1<a href="$2" target="_blank">$2</a>', $tweet);
					$tweet = preg_replace('/(^|[^a-zA-Z0-9\p{L}])@([a-zA-Z0-9_\p{L}]+)/im', '$1<a href="https://twitter.com/$2" target="_blank">@$2</a>', $tweet);
					?>
					<tr<?= $alt ? ' class="alt"' : '' ?>>
						<td>
							<?= $tweet ?>
							<small class="subtle"><a href="https://twitter.com/<?= $_['poe']['twitter_name'] ?>/status/<?= htmlspecialchars($row['twitter_tweet_id']) ?>" class="js-time" data-time="<?= htmlspecialchars($row['tweet_time']) ?>"><?= create_date($row['tweet_time']) ?></a></small> 
						</td>
					</tr>
					<?php
					$alt = !$alt;
				}
				?>
			</table>
			<div class="right"><small><a href="https://twitter.com/<?= $_['poe']['twitter_name'] ?>">@pathofexile</a></small></div>
		</div>
	</div>
	
	<div class="content-box">
		<h1>Forums</h1>
		<a href="rss.php"><img src="images/rss-icon-28.png" class="rss-icon" /></a>
		<table class="list">
			<tr>
				<th style="width: 30px;">#</th>
				<th>Thread</th>
				<th>Poster</th>
				<th style="width: 200px;">Time</th>
				<th>Forum</th>
			</tr>
		
			<?php
			$result = $SQL->query("
				SELECT p.*, b.post_count 
				FROM posts p 
				JOIN (
						SELECT MAX(forum_post_id) max_forum_post_id, COUNT(post_id) post_count 
						FROM posts 
						GROUP BY forum_thread_id 
						ORDER BY max_forum_post_id DESC 
						LIMIT ".($page - 1) * $threads_per_page.", ".$threads_per_page."
					) b 
				ON b.max_forum_post_id = p.forum_post_id
			");
			
			$alt = false;
			while ($row = $SQL->fetch_assoc($result)) {
				?>
				<tr<?= $alt ? ' class="alt"' : '' ?>>
					<td class="center"><a href="http://www.pathofexile.com/forum/view-thread/<?= htmlspecialchars($row['forum_thread_id']) ?>/filter-account-type/staff"><?= $row['post_count'] ?></a></td>
					<td><a href="http://www.pathofexile.com/forum/view-thread/<?= htmlspecialchars($row['forum_thread_id']) ?>/page/<?= htmlspecialchars($row['forum_page_number']) ?>#p<?= htmlspecialchars($row['forum_post_id']) ?>"><?= htmlspecialchars($row['thread_name']) ?></a></td>
					<td class="center"><a class="ggg" href="http://www.pathofexile.com/account/view-profile/<?= htmlspecialchars($row['post_author']) ?>"><?= htmlspecialchars($row['post_author']) ?></a></td>
					<td class="center js-time" data-time="<?= htmlspecialchars($row['post_time']) ?>"><?= create_date($row['post_time']) ?></td>
					<td class="center"><a href="http://www.pathofexile.com/forum/view-forum/<?= htmlspecialchars($row['forum_forum_id']) ?>"><?= htmlspecialchars($row['forum_name']) ?></a></td>
				</tr>
				<?php
				$alt = !$alt;
			}
			?>
		
		</table>
		<div class="right">
			<small><?= $page_links ?></small>
		</div>
	</div>	
</center>

<?php
require_once "./includes/footer.inc.php";
?>
