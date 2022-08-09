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
    <link rel="stylesheet" type="text/css" href="static/style.css?v5" />
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
			<ul id="locale-selection">
				{{range .Locales}}
				<li{{if eq .Subdomain $.Locale.Subdomain}} class="selected-locale"{{end}}><a href="{{call $.SubdomainURL .Subdomain}}"><img src="{{.Image}}" /></a></li>
				{{end}}
			</ul>
		</header>
        <div class="content-box">
            <h1>{{call $.Translate "Activity"}}</h1>
            <a href="rss"><img src="static/images/rss-icon-28.png" class="rss-icon" /></a>
            <table id="activity-table" class="list">
                <thead>
                    <tr>
                        <th></th>
                        <th></th>
                        <th>{{call $.Translate "Thread"}}</th>
                        <th>{{call $.Translate "Poster"}}</th>
                        <th>{{call $.Translate "Time"}}</th>
                        <th>{{call $.Translate "Forum"}}</th>
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

    <script src="static/index.js?v7"></script>
</body>
</html>`

type IndexConfiguration struct {
	GoogleAnalytics string
}

var indexTemplate *template.Template

func init() {
	t, err := template.New("index").Parse(index)
	if err != nil {
		panic(err)
	}
	indexTemplate = t
}

func IndexHandler(configuration IndexConfiguration) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		locale := LocaleForRequest(c.Request())
		return indexTemplate.Execute(c.Response(), struct {
			Configuration IndexConfiguration
			Locales       []*Locale
			Locale        *Locale
			Translate     func(string) string
			SubdomainURL  func(string) string
		}{
			Configuration: configuration,
			Locales:       Locales,
			Locale:        locale,
			Translate:     locale.Translate,
			SubdomainURL: func(subdomain string) string {
				return SubdomainURL(c, subdomain)
			},
		})
	}
}
