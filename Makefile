app/km: app/km.go
	go build -o app/km app/km.go app/kilometers.go app/errors.go app/db.go

.PHONY: test-run
test-run: app/km
	-pkill km
	cp ./config-testing.yml app/config.yml
	cd app && ./km &

.PHONY: start-postgres
start-postgres:
	docker kill `cat .cidfile`
	rm .cidfile
	docker run -cidfile=./.cidfile -v /home/fkalter/postgresdata:/data:rw\
				       -v /home/fkalter/github/km/log:/log:rw\
				       -d -p 5432:5432\
				       freekkalter/postgres-supervisord:km /usr/bin/supervisord

.PHONY: prepare-production
prepare-production: app/km minify
	cp config-production.yml app/config.yml

.PHONY:
build-production: Dockerfile
	docker build -t freekkalter/km:deploy .

.PHONY: run-local-production
run-local-production: prepare-production
	-pkill km
	-docker kill `cat .cidfile`
	-rm .cidfile
	-docker rm km_production
	docker run -name km_production -cidfile=./.cidfile -v /home/fkalter/postgresdata:/data:rw\
				       -v /home/fkalter/github/km/log:/log:rw\
					   -d -p 4001:4001\
				       freekkalter/km:deploy /usr/bin/supervisord
	-docker kill `cat .nginx.cidfile`
	-rm .nginx.cidfile
	docker run -d -p 443:443 -link km_production:app -cidfile=./.nginx.cidfile \
			   -v /home/fkalter/nginx/ssl:/etc/nginx/conf:ro \
			   freekkalter/nginx:start_nginx /start_nginx

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

# target: clean - Removes minified CSS and JS files.
.PHONY: clean
clean:
	rm -f $(CSS_MINIFIED)
	rm -f app/km
	rm -f app/js/*.min.js
	rm -f app/js/combined*.js
