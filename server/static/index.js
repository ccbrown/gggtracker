var currentPage = undefined;
var currentHideHelp = undefined;

var POE = {
    Forum: {
        SpoilerClick: function(element) {
            var $spoiler = $(element).closest('.spoiler');
            if ($spoiler.hasClass('spoilerHidden')) {
                $(element).val("Hide");
                $spoiler.removeClass('spoilerHidden');
            } else {
                $(element).val("Show");
                $spoiler.addClass('spoilerHidden');
            }
        },
    },
};

function loadActivity() {
    var params = new URLSearchParams(location.hash.replace(/^#/, ''));
    var page = params.get('page') || '';
    var hideHelp = params.has('nohelp') && params.get('nohelp') !== 'false';
    if (currentPage !== undefined && page == currentPage && currentHideHelp !== undefined && hideHelp == currentHideHelp) {
        return;
    }
    var previousPage = currentPage;
    var previousHideHelp = currentHideHelp;

    currentPage = page;
    currentHideHelp = hideHelp;

    var canonicalNohelpParam = hideHelp ? '&nohelp' : '';

    $('#activity-table tbody').empty().append($('<tr>').append($('<td>').attr('colspan', 6).text('Loading...')))

    $.get('activity.json?next=' + page + canonicalNohelpParam, function(data) {
        var $tbody = $('#activity-table tbody');
        $tbody.empty();

        for (var i = 0; i < data.activity.length; ++i) {
            var type = data.activity[i].type;
            var activity = data.activity[i].data;
            var subreddit = activity.subreddit || 'pathofexile';
            var $tr = $('<tr>').addClass(type == 'forum_post' ? 'forum' : 'reddit').addClass(type.replace('/_/', '-'));

            var $toggleTD = $('<td class="toggle">');
            $tr.append($toggleTD);

            if (type == 'forum_post') {
                $tr.append($('<td class="icon">').append($('<a>')
                    .attr('href', 'https://' + activity.host + '/forum/view-thread/' + activity.thread_id + '/filter-account-type/staff')
                    .append($('<img src="static/images/forum-thread.png" />'))
                ));
            } else if (type == 'reddit_comment') {
                $tr.append($('<td class="icon">').append($('<a>')
                    .attr('href', 'https://www.reddit.com/r/' + subreddit + '/comments/' + activity.post_id)
                    .append($('<img src="static/images/snoo.png" />'))
                ));
            } else {
                $tr.append($('<td class="icon">').append($('<a>')
                    .attr('href', 'https://www.reddit.com' + activity.permalink)
                    .append($('<img src="static/images/snoo.png" />'))
                ));
            }

            if (type == 'forum_post') {
                $tr.append($('<td class="title">').append($('<a>')
                    .attr('href', 'https://' + activity.host + '/forum/view-post/' + activity.id)
                    .text(activity.thread_title)
                ));
            } else if (type == "reddit_post") {
                $tr.append($('<td class="title">').append($('<a>')
                    .attr('href', activity.url ? activity.url : ('https://www.reddit.com' + activity.permalink))
                    .text(activity.title)
                ));
            } else if (type == "reddit_comment") {
                $tr.append($('<td class="title">').append($('<a>')
                    .attr('href', 'https://www.reddit.com/r/' + subreddit + '/comments/' + activity.post_id + '/-/' + activity.id + '/?context=3')
                    .text(activity.post_title)
                ));
            }

            if (type == 'forum_post') {
                $tr.append($('<td class="poster">').append($('<a>')
                    .attr('href', 'https://' + activity.host + '/account/view-profile/' + encodeURIComponent(activity.poster) + '-' + String(activity.poster_discriminator || 0).padStart(4, '0'))
                    .text(activity.poster)
                ));
            } else {
                $tr.append($('<td class="poster">').append($('<a>')
                    .attr('href', 'https://www.reddit.com/user/' + encodeURIComponent(activity.author))
                    .text(activity.author)
                ));
            }

            $tr.append($('<td class="time">').text((new Date(Date.parse(activity.time))).toLocaleString()));

            if (type == 'forum_post') {
                $tr.append($('<td class="forum">').append($('<a>')
                    .attr('href', 'https://' + activity.host + '/forum/view-forum/' + encodeURIComponent(activity.forum_id))
                    .text(activity.forum_name)
                ));
            } else {
                $tr.append($('<td class="forum">').append($('<a>')
                    .attr('href', 'https://www.reddit.com/r/' + subreddit)
                    .text(subreddit)
                ));
            }

            $tbody.append($tr);

            if (!activity.body_html) { continue; }

            $tr = $('<tr>').addClass(type == 'forum_post' ? 'forum' : 'reddit').hide();
            var $body = $('<td colspan="6" class="body">');
            $tr.append($body);

            $body.html(activity.body_html);
            $body.find('a').each(function() {
                var r = $(this).attr('href');
                if (r && (r.indexOf(':') < 0 || r.indexOf('/') <= r.indexOf(':'))) {
                    var root = type == 'forum_post' ? 'https://' + activity.host : 'https://www.reddit.com';
                    $(this).attr('href', root + (r[0] == '/' ? '' : '/') + r);
                }
            });

            var $expander = $('<img class="expander" src="static/images/expand.svg" />');
            var $collapser = $('<img class="collapser" src="static/images/collapse.svg" />').hide();
            $expander.data('collapser', $collapser).data('body', $tr);
            $collapser.data('expander', $expander).data('body', $tr);

            $expander.click(function() {
                $(this).hide();
                $(this).data('collapser').show();
                $(this).data('body').show();
            });

            $collapser.click(function() {
                $(this).hide();
                $(this).data('expander').show();
                $(this).data('body').hide();
            });

            $toggleTD.append($expander).append($collapser);

            $tbody.append($tr);
        }

        if (hideHelp) {
            $('#hide-help-forum').hide();
            $('#show-help-forum').attr('href', page ? '#page=' + page : '#').show();
        } else {
            $('#show-help-forum').hide();
            $('#hide-help-forum').attr('href', (page ? '#page=' + page + '&' : '#') + 'nohelp').show();
        }

        $('#activity-nav').empty().append($('<a>').text('Next Page').attr('href', '#page=' + data.next + canonicalNohelpParam).click(function() {
            window.scrollTo(0, 0);
        }));
    }).fail(function() {
        alert('Something went wrong. Better luck next time.');
        currentPage = previousPage
        if (currentPage !== undefined) {
            var previousNohelpParam = previousHideHelp ? '&nohelp' : '';
            window.location.hash = 'page=' + currentPage + previousNohelpParam;
        } else {
            window.location.hash = '';
        }
    })
}

$(function() {
    $(window).on('hashchange', function() {
        loadActivity();
    });
    loadActivity();
})
