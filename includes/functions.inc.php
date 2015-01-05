<?php
if (!defined('IN_GGGT')) { die(); }

function select_options($options, $selected = '') {
	$ret = '';
	if (is_array($options)) {
		if (!isset($options[$selected])) {
			$ret .= '<option value="'.htmlspecialchars($selected).'" selected=true> </option>';
		}
		foreach ($options as $value => $text) {
			$ret .= '<option value="'.htmlspecialchars($value).'"'.($value == $selected ? ' selected=true' : '').'>'.htmlspecialchars($text).'</option>';
		}
	}
	return $ret;
}

function error_handler($errno, $errstr, $errfile, $errline) {
	if ($errno == E_USER_ERROR) {
		die('<title>Error!</title><b>Error:</b> A fatal error has occured. Please try again later as the problem will be fixed shortly.');
	}
}

function selected($var, $option_value) {
	return (isset($_POST[$var]) && $_POST[$var] == $option_value ? ' selected=true' : '');
}

function checked($var, $option_value) {
	return (isset($_POST[$var]) && $_POST[$var] == $option_value ? ' checked=true' : '');
}

function field_value($post_var, $var2 = '') {
	if (isset($_POST[$post_var])) {
		return htmlspecialchars($_POST[$post_var]);
	} else {
		return $var2;
	}
}

function fetch_isp() {
	$ip_addr = fetch_ip_addr();
	$domain_name = $ip_addr ? @gethostbyaddr($ip_addr) : '';
	$isp_name = explode('.', $domain_name);
	$isp_name = array_slice($isp_name, count($isp_name)-2);
	$isp_name = implode('.', $isp_name);
	if (!preg_match('#[a-zA-Z]+#', $isp_name)) {
		$isp_name = "Unknown";
	}
	return $isp_name;
}

function fetch_ip_addr() {
	$ip_addr = $_SERVER['REMOTE_ADDR'];
	
	if (strpos($ip_addr, ',')) {
	    $ip_addr = explode(',', $ip_addr);
	    $ip_addr = trim($ip_addr[0]);
	}
	
	return $ip_addr;
}

function build_simple_link($base_name, $link_title, $query) {
	return sprintf('<a href="%s">%s</a>', $base_name . '?' . $query, $link_title);
}

function create_date($timestamp, $format = NULL) {
	global $_;
	
	if ($timestamp === NULL) {
		$date = 'N/A';
	} else {
		$date = gmdate($format === NULL ? $_['page']['php_date_format'] : $format, $timestamp + $_['page']['php_time_offset']);
	}
	
	return $date;
}

function create_gmdate($timestamp, $format = NULL) {
	global $_;
	
	if ($timestamp === NULL) {
		$date = 'N/A';
	} else {
		$date = gmdate($format === NULL ? $_['page']['php_date_format'] : $format, $timestamp);
	}
	
	return $date;
}

function unhtmlspecialchars($string) {
	$string = str_replace('&quot;', '"', $string);
	$string = str_replace('&lt;', '<', $string);
	$string = str_replace('&gt;', '>', $string);
	$string = str_replace('&amp;', '&', $string);
	
	return $string;
}

function stripslashes_deep(&$value) { 
    $value = is_array($value) ? 
                array_map('stripslashes_deep', $value) : 
                stripslashes($value); 

    return $value; 
}
?>