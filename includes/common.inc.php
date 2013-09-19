<?php
if (!defined('IN_GGGT')) { die(); }

error_reporting(E_ALL);

ob_start('ob_gzhandler');

$_ = require_once('config.inc.php');

$_['page']['start_time'] = array_sum(explode(' ', microtime()));
$_['page']['time']       = time();

require_once 'functions.inc.php';
require_once 'functions_ggg.inc.php';
require_once 'classes/SQL.class.php';

set_error_handler('error_handler');

if ((function_exists("get_magic_quotes_gpc") && get_magic_quotes_gpc()) || ini_get('magic_quotes_sybase')) {
	stripslashes_deep($_GET);
	stripslashes_deep($_POST);
	stripslashes_deep($_COOKIE);
	stripslashes_deep($_SESSION);
	stripslashes_deep($_REQUEST);
}

$SQL = new SQL($_['mysql']['host'], $_['mysql']['username'], $_['mysql']['password'], $_['mysql']['pconnect']);
$SQL->select_db($_['mysql']['db_name']);
?>