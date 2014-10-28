app/km: app/km.go
	go build -o app/km ./app

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

# target: minify-js - Minifies JS.
minify-js: app/js/master.js

app/js/angular-combined.js: app/js/app.js app/js/controller.js  app/js/animations.js
	cat app/js/app.js app/js/controller.js app/js/animations.js > app/js/angular-combined.js

app/js/angular-combined.anno.js: app/js/angular-combined.js
	ngmin app/js/angular-combined.js app/js/angular-combined.anno.js

app/js/angular-combined.anno.min.js: app/js/angular-combined.anno.js
	-$(YUI_COMPRESSOR) $(YUI_COMPRESSOR_FLAGS) --type js app/js/angular-combined.anno.js > app/js/angular-combined.anno.min.js

app/js/master.js: app/js/angular-combined.anno.min.js
	cat app/js/jquery.min.js\
		app/js/angular.min.js\
		app/js/angular-route.min.js\
		app/js/angular-animate.min.js\
		app/js/ui-bootstrap-custom-tpls-0.7.0.Minimale.min.js\
		app/js/bootstrap-datepicker.js\
		app/js/angular-combined.anno.min.js\
		| sed '/^\/\//d' > app/js/master.js


# target: clean - Removes minified CSS and JS files.
.PHONY: clean
clean:
	rm -f $(CSS_MINIFIED)
	rm -f app/km
	rm -f app/js/angular-combined*
	rm -f app/js/master.js
