<?php
if (!defined('IN_GGGT')) { die(); }

$config = array(

	'page'   => array(
		/**
		* The format that dates constructed via PHP are displayed in. See
		* http://php.net/manual/en/function.date.php for more information.
		*/
		'php_date_format' => 'M j, Y G:i A',

		/**
		* The offset in seconds of the default time zone used for displaying dates.
		*/
		'php_time_offset' => 0,

		/**
		* The format that dates constructed via JavaScript are displayed in. See
		* https://github.com/phstc/jquery-dateFormat for more information.
		*
		* This is used when possible to ensure that dates match the visitor's local
		* time zone. It should be equivalent to the PHP date format.
		*/
		'js_date_format' => 'MMM d, yyyy h:mm a',

		/**
		* Footer HTML.
		*/
		'footer' =>
			'Please direct feedback to <a href="http://www.pathofexile.com/forum/view-thread/69448" target="_blank">this thread</a>. '.
			'Want a new feature? <a href="https://github.com/ccbrown/gggtracker" target="_blank">Add it yourself!</a>',
	),

	/**
	* MySQL information. Don't forget to initialize the database with the contents of database.sql.
	*/
	'mysql' => array(
		'host'      => '',
		'username'  => '',
		'password'  => '',
		'db_name'   => '',
		'pconnect'  => false,
	),

	/**
	* To poll forum posts and Twitter updates, two cron jobs should be set up to invoke
	* scan_posts.cron.php and update_tweets.cron.php. These are publically accessible
	* files, so to prevent abuse, you must create a random secret here. Then, when
	* invoking the scripts, pass that secret as a GET parameter.
	*
	* Make this long and random.
	*/
	'cron' => array(
		'secret' => '',
	),

	'poe' => array(
		/**
		* The timezone of the forum account used (e.g. 'GMT' or 'PST').
		*/
		'forum_timezone' => '',

		/**
		* The value of the PHPSESSID cookie for a valid login session.
		*/
		'forum_sessid' => '',

		/**
		* The id of the news forum.
		*/
		'news_forum_id' => 54,

		/**
		* An array of forum posters to track.
		*/
		'forum_posters' => array(
			'Chris', 'Jonathan', 'Erik', 'Mark_GGG', 'Samantha', 'Rory',
			'Rhys', 'Qarl', 'Andrew_GGG', 'Damien_GGG', 'Russell', 'Joel_GGG', 'Ari',
			'Thomas', 'BrianWeissman', 'Edwin_GGG', 'Support', 'Dylan', 'MaxS',
			'Ammon_GGG', 'Jess_GGG', 'Robbie_GGG', 'GGG_Neon', 'Jason_GGG', 'Henry_GGG',
			'Michael_GGG', 'Bex_GGG', 'Cagan_GGG', 'Daniel_GGG', 'Kieren_GGG', 'Yeran_GGG',
			'Gary_GGG', 'Dan_GGG', 'Jared_GGG', 'Brian_GGG', 'RobbieL_GGG', 'Arthur_GGG',
			'NickK_GGG', 'Felipe_GGG', 'Alex_GGG', 'Alexcc_GGG', 'Andy', 'CJ_GGG',
			'Eben_GGG', 'Emma_GGG', 'Ethan_GGG', 'Fitzy_GGG', 'Hartlin_GGG', 'Jake_GGG',
			'Lionel_GGG', 'Melissa_GGG', 'MikeP_GGG', 'Novynn', 'Rachel_GGG', 'Rob_GGG',
			'Roman_GGG', 'Sarah_GGG', 'SarahB_GGG', 'Tom_GGG'
		),

		/**
		* The Twitter user to track.
		*/
		'twitter_name' => 'pathofexile',
	),

	/**
	* Twitter-supplied credentials. Create an app at https://apps.twitter.com and obtain
	* these through the "Keys and Access Tokens" tab for your app.
	*/
	'twitter' => array(
		'consumer_key'    => '',
		'consumer_secret' => '',
		'access_token'    => '',
		'access_secret'   => '',
	),
);

/**
* Any of the above can be overridden in a local configuration file. Simply create a file named
* local_config.php that looks like this:
*
* <?php
* return array(
*     '[thing you want to override]' => [the value you want to use],
*     ...
* );
* ?>
*
*/
$local = @include('local_config.inc.php');

if (isset($local['date_format']) || isset($local['time_offset'])) {
	die('out-dated configuration. review the changes to config.inc.php');
}

return $local ? array_replace_recursive($config, $local) : $config;
?>
