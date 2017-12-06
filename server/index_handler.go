package server

import (
	"html/template"

	"github.com/labstack/echo"
)

var index = `<!DOCTYPE html><html>
<head>
    <meta http-equiv="Content-Type" content="text/html;charset=utf-8" />
    <title>GGG Tracker</title>
    <link rel="shortcut icon" href="static/favicon.ico" />
    <link rel="stylesheet" type="text/css" href="static/style.css?v4" />
    <link rel="alternate" type="application/rss+xml" title="GGG Tracker Forum Feed" href="rss" />

    <script src="https://ajax.googleapis.com/ajax/libs/jquery/3.1.0/jquery.min.js"></script>

	{{if .Configuration.GoogleAnalytics}}
    <script type="text/javascript">
    var _gaq = _gaq || [];
    _gaq.push(['_setAccount', '{{.Configuration.GoogleAnalytics}}']);
    _gaq.push(['_trackPageview']);

    (function() {
        var ga = document.createElement('script'); ga.type = 'text/javascript'; ga.async = true;
        ga.src = ('https:' == document.location.protocol ? 'https://ssl' : 'http://www') + '.google-analytics.com/ga.js';
        var s = document.getElementsByTagName('script')[0]; s.parentNode.insertBefore(ga, s);
    })();
    </script>
	{{end}}
</head>
<body>
    <div class="container">
		<header>
			<a href="/"><img src="static/images/ggg-dark.png" /></a>
		</header>
        <div class="content-box">
            <h1>Activity</h1>
            <a href="rss"><img src="static/images/rss-icon-28.png" class="rss-icon" /></a>
            <table id="activity-table" class="list">
                <thead>
                    <tr>
                        <th></th>
                        <th></th>
                        <th>Thread</th>
                        <th>Poster</th>
                        <th>Time</th>
                        <th>Forum</th>
                    </tr>
                </thead>
                <tbody>
                </tbody>
            </table>
            <div id="activity-nav" class="right"></div>
        </div>
        <footer>
            <p>This site is not affiliated with Path of Exile or Grinding Gear Games in any way.</p>
			<p>
				Please direct feedback to <a href="https://www.pathofexile.com/forum/view-thread/69448" target="_blank">this thread</a>.
				Want a new feature? <a href="https://github.com/ccbrown/gggtracker" target="_blank">Add it yourself!</a>
			</p>
        </footer>
    </div>

    <script src="static/index.js?v5"></script>
</body>
</html>`

type IndexConfiguration struct {
	GoogleAnalytics string
}

func IndexHandler(configuration IndexConfiguration) echo.HandlerFunc {
	t, err := template.New("index").Parse(index)
	if err != nil {
		panic(err)
	}
	return func(c echo.Context) error {
		return t.Execute(c.Response(), struct {
			Configuration IndexConfiguration
		}{
			Configuration: configuration,
		})
	}
}
