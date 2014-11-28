# Patterns matching CSS files that should be minified. Files with a .min.css
# suffix will be ignored.
CSS_FILES = $(filter-out %.min.css,$(wildcard \
	static/css/*.css \
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
minify-js: static/js/master.js

static/js/angular-combined.js: static/js/app.js static/js/controller.js  static/js/animations.js
	cat static/js/app.js static/js/controller.js static/js/animations.js > static/js/angular-combined.js

static/js/angular-combined.anno.js: static/js/angular-combined.js
	ngmin static/js/angular-combined.js static/js/angular-combined.anno.js

static/js/angular-combined.anno.min.js: static/js/angular-combined.anno.js
	-$(YUI_COMPRESSOR) $(YUI_COMPRESSOR_FLAGS) --type js static/js/angular-combined.anno.js > static/js/angular-combined.anno.min.js

static/js/master.js: static/js/angular-combined.anno.min.js
	cat static/js/jquery.min.js\
		static/js/angular.min.js\
		static/js/angular-route.min.js\
		static/js/angular-animate.min.js\
		static/js/ui-bootstrap-custom-tpls-0.7.0.Minimale.min.js\
		static/js/bootstrap-datepicker.js\
		static/js/angular-combined.anno.min.js\
		| sed '/^\/\//d' > static/js/master.js


# target: clean - Removes minified CSS and JS files.
.PHONY: clean
clean:
	rm -f $(CSS_MINIFIED)
	rm -f static/km
	rm -f static/js/angular-combined*
	rm -f static/js/master.js
