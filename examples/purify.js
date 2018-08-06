/*

Functions are from the DOMPurify examples at https://github.com/cure53/DOMPurify/blob/master/demos/README.md

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.

*/

const registerSanitizeAllLinksToExternal = () => {
	DOMPurify.addHook('afterSanitizeAttributes', function(node) {
		// set all elements owning target to target=_blank
		if ('target' in node) {
			node.setAttribute('target', '_blank');
		}
		// set non-HTML/MathML links to xlink:show=new
		if (!node.hasAttribute('target')
			&& (node.hasAttribute('xlink:href')
				|| node.hasAttribute('href'))) {
			node.setAttribute('xlink:show', 'new');
		}
	});
}
registerSanitizeAllLinksToExternal();

const registerSanitizeHTTPLeaks = () => {
	// Specify attributes to proxy
	var attributes = ['action', 'background', 'href', 'poster', 'src'];

	// specify the regex to detect external content
	var regex = /(url\("?)(?!data:)/gim;

	/**
	 *  Take CSS property-value pairs and proxy URLs in values,
	 *  then add the styles to an array of property-value pairs
	 */
	function addStyles(output, styles) {
		for (var prop = styles.length-1; prop >= 0; prop--) {
			if (styles[styles[prop]]) {
				var url = styles[styles[prop]].replace(regex, '$1' + proxy);
				styles[styles[prop]] = url;
			}
			if (styles[styles[prop]]) {
				output.push(styles[prop] + ':' + styles[styles[prop]] + ';');
			}
		}
	}

	/**
	 * Take CSS rules and analyze them, proxy URLs via addStyles(),
	 * then create matching CSS text for later application to the DOM
	 */
	function addCSSRules(output, cssRules) {
		for (var index = cssRules.length-1; index >= 0; index--) {
			var rule = cssRules[index];
			// check for rules with selector
			if (rule.type == 1 && rule.selectorText) {
				output.push(rule.selectorText + '{')
				if (rule.style) {
					addStyles(output, rule.style)
				}
				output.push('}');
			// check for @media rules
			} else if (rule.type === rule.MEDIA_RULE) {
				output.push('@media ' + rule.media.mediaText + '{');
				addCSSRules(output, rule.cssRules)
				output.push('}');
			// check for @font-face rules
			} else if (rule.type === rule.FONT_FACE_RULE) {
				output.push('@font-face {');
				if (rule.style) {
					addStyles(output, rule.style)
				}
				output.push('}');
			// check for @keyframes rules
			} else if (rule.type === rule.KEYFRAMES_RULE) {
				output.push('@keyframes ' + rule.name + '{');
				for (var i=rule.cssRules.length-1; i>=0; i--) {
					var frame = rule.cssRules[i];
					if (frame.type === 8 && frame.keyText) {
						output.push(frame.keyText + '{');
						if (frame.style) {
							addStyles(output, frame.style);
						}
						output.push('}');
					}
				}
				output.push('}');
			}
		}
	}

	/**
	 * Proxy a URL in case it's not a Data URI
	 */
	function proxyAttribute(url) {
		if (/^data:image\//.test(url)) {
			return url;
		} else {
			console.debug('external resource needs proxy', url);
			return '//api/safe-proxy/fetch/?uri=' + escape(url);
		}
	}

	// Add a hook to enforce proxy for leaky CSS rules
	DOMPurify.addHook('uponSanitizeElement', function(node, data) {
		if (data.tagName === 'style') {
			var output  = [];
			addCSSRules(output, node.sheet.cssRules);
			node.textContent = output.join("\n");
		}
	});

	// Add a hook to enforce proxy for all HTTP leaks incl. inline CSS
	DOMPurify.addHook('afterSanitizeAttributes', function(node) {

		// Check all src attributes and proxy them
		for(var i = 0; i <= attributes.length-1; i++) {
			if (node.hasAttribute(attributes[i])) {
				node.setAttribute(attributes[i], proxyAttribute(
					node.getAttribute(attributes[i]))
				);
			}
		}

		// Check all style attribute values and proxy them
		if (node.hasAttribute('style')) {
			var styles = node.style;
			var output = [];
			for(var prop = styles.length-1; prop >= 0; prop--) {
				// we re-write each property-value pair to remove invalid CSS
				if (node.style[styles[prop]] && regex.test(node.style[styles[prop]])) {
					var url = node.style[styles[prop]].replace(regex, '$1' + proxy)
					node.style[styles[prop]] = url;
				}
				output.push(styles[prop] + ':' + node.style[styles[prop]] + ';');
			}
			// re-add styles in case any are left
			if (output.length) {
				node.setAttribute('style', output.join(""));
			} else {
				node.removeAttribute('style');
			}
		}
	});
};
registerSanitizeHTTPLeaks();
