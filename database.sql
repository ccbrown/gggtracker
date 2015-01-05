CREATE TABLE `posts` (
  `post_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `forum_thread_id` int(11) unsigned NOT NULL,
  `forum_forum_id` int(11) unsigned NOT NULL,
  `post_time` int(11) NOT NULL,
  `post_author` tinytext NOT NULL,
  `post_body` mediumtext NOT NULL,
  `forum_post_id` int(11) unsigned NOT NULL,
  `thread_name` tinytext NOT NULL,
  `forum_name` tinytext NOT NULL,
  `forum_page_number` int(11) NOT NULL,
  PRIMARY KEY (`post_id`),
  KEY `forum_post_id` (`forum_post_id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;



CREATE TABLE `tweets` (
  `tweet_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `twitter_tweet_id` bigint(21) NOT NULL,
  `tweet_text` tinytext NOT NULL,
  `tweet_time` int(11) NOT NULL,
  PRIMARY KEY (`tweet_id`),
  KEY `twitter_tweet_id` (`twitter_tweet_id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;




/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
