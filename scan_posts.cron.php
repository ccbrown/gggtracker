<?php
define('IN_GGGT', true);
require_once "./includes/common.inc.php";

if (!isset($_GET['secret']) || $_GET['secret'] != $_['cron']['secret']) {
	die();
}

foreach ($_['poe']['forum_posters'] as $poster) {
	echo 'Indexing posts by '.htmlspecialchars($poster).'...<br />';
	if (!index_poster($poster)) {
		echo 'Failure...<br />';
	}
}
?>
