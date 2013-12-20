# Patterns matching CSS files that should be minified. Files with a .min.css
# suffix will be ignored.
CSS_FILES = $(filter-out %.min.css,$(wildcard \
	app/css/*.css \
))

# Command to run to execute the YUI Compressor.
YUI_COMPRESSOR = /usr/bin/yui-compressor

# Flags to pass to the YUI Compressor for both CSS and JS.
YUI_COMPRESSOR_FLAGS = --charset utf-8 --verbose

CSS_MINIFIED = $(CSS_FILES:.css=.min.css)
production: app/km minify production-config

production-config:
	cp config-production.yml app/config.yml

testing: app/km minify test-run

app/km: app/km.go
	go build -o app/km app/km.go

# target: minify - Minifies CSS and JS.
minify: minify-css minify-js

# target: minify-css - Minifies CSS.
minify-css: $(CSS_FILES) $(CSS_MINIFIED)

%.min.css: %.css
	@echo '==> Minifying $<'
	$(YUI_COMPRESSOR) $(YUI_COMPRESSOR_FLAGS) --type css $< >$@
	@echo

# target: minify-js - Minifies JS.
minify-js: app/js/combined.anno.js app/js/combined.anno.min.js

app/js/combined.js: app/js/app.js app/js/controller.js
	cat app/js/app.js app/js/controller.js app/js/animations.js > app/js/combined.js

app/js/combined.anno.js: app/js/combined.js
	ngmin app/js/combined.js app/js/combined.anno.js

app/js/combined.anno.min.js: app/js/combined.anno.js
	$(YUI_COMPRESSOR) $(YUI_COMPRESSOR_FLAGS) --type js app/js/combined.anno.js > app/js/combined.anno.min.js

# TODO: find a way to keep the container running and just restart the app
# already running supervisor, so maybe also start a service wich listens for restart commands (might be overkill)
test-run: app/km
	docker kill `cat .cidfile`
	rm .cidfile
	cp config-testing.yml app/config.yml
	docker run -cidfile=./.cidfile -v /home/fkalter/postgresdata:/data:rw\
								   -v /home/fkalter/github/km/app:/app:rw\
								   -v /home/fkalter/github/km/log:/log:rw\
								   -d -p 4001:4001 -p 5432:5432\
								   freekkalter/postgres-supervisord:km /usr/bin/supervisord

test-deploy: app/km
	docker build freekkalter/km:deploy .
	docker kill `cat .cidfile`
	rm .cidfile
	cp config-production.yml app/config.yml
	docker run -cidfile=./.cidfile -v /home/fkalter/postgresdata:/data:rw\
								   -v /home/fkalter/github/km/app:/app:rw\
								   -v /home/fkalter/github/km/log:/log:rw\
								   -d -p 4001:4001 -p 5432:5432\
								   freekkalter/km:deploy /usr/bin/supervisord

# target: clean - Removes minified CSS and JS files.
clean:
	rm -f $(CSS_MINIFIED)
	rm -f app/km
