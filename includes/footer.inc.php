<?php
if (!defined('IN_GGGT')) { die(); }
?>
				<script>
				$(function() {
					$('.js-time').each(function() {
						$(this).text($.format.date(new Date($(this).data('time') * 1000), 'MMM d, yyyy h:mm a'));
					});
				});
				</script>
		
				<div class="footer">
					Page generated in <?= round(array_sum(explode(' ', microtime())) - $_['page']['start_time'], 6) ?> seconds. 
					This site is not affiliated with Path of Exile or Grinding Gear Games in any way. <br />
					Please direct feedback to <a href="http://www.pathofexile.com/forum/view-thread/69448" target="_blank">this thread</a>.
				</div>
			</div>
		</center>
	</body>
</html>