package main

import (
    "fmt"
    "io"
)

func dumpCoreJS(out io.Writer) {
    fmt.Fprintln(out, `
/*************************************************************
 *  Twine-like Text Game IO interface
 * 
 *  Version 1.1
 *
 *************************************************************/


/*
 * CLASSES:
 * 
 * .io-splash  (h1 splash, h2 subtitle, h3 author)
 * .io-clear   (hr when clearing))
 * .io-title   (t)
 * 
 */

//var $ = require("jquery-browserify");

const $ = (name) => document.querySelector(name)

var _id = "#play";    // the main window

var _addr = 0;



function clean (text) { 
    var t = text;
    t = t.replace(/<q>/g,"\"").replace(/<\/q>/g,"\"");
    // remove tags
    t = t.replace(/<[^>]*>/g,"");
    t = t.replace(/&lt;/g,"<")
	.replace(/&gt;/g,">")
	.replace(/&nbsp;/g," ")
	.replace(/&amp;/g,"&");
    t = t.replace(/&trade;/g,"(TM)")
	.replace(/&mdash;/g," -- ");
    t = t.replace(/&bull;/g,"*");
    // by default, get rid of special characters
    t = t.replace(/&[^;]*;/g,"");
    return t;
}

var _seed = 0;

function fresh_id () { 
    _seed++;
    return "_freshid"+_seed;
}


var CLEAR_FLAG = false;   // every newp automatically clears the screen 
var DEBUG_FLAG = false;   // create debug information



function config (c) {
    CLEAR_FLAG = c.clear ? true : false;
    DEBUG_FLAG = c.debug ? true : false;
}

function paragraphs (text) {
    var t = text;
    return t.split(/\n[ \t]*\n/);
}

function ce (s) { 
    return document.createElement(s);
}



// function that can be used for scrolling behavior 

function scroll (addr) {
    // fast:
    //$(document).scrollTop($("#io-addr"+addr).offset().top);
    var tp = $("#io-addr"+addr).offset().top;
    $("html,body").animate({scrollTop:tp},200);
}


function splash (txt,subtxt,author) { 
    _addr = 0;
    var content = "<h1 class=\"io-splash\" id=\"io-addr0\">"+txt+"</h1>";
    if (subtxt) { 
	content += "<h2 class=\"io-splash\">"+subtxt+"</h2>";
    }
    if (author) { 
	content += "<h3 class=\"io-splash\">By "+author+"</h3>";
    }
    $(_id).innerHTML = content;
    return this;
}


function newp (cl_flag) {
    // new passage

    if (CLEAR_FLAG || cl_flag) {
	
	$(_id).innerHTML = '';
	return this;
    }
    
    _addr += 1;
    const hr = ce("hr");
    hr.classList.add("io-clear")
    hr.setAttribute("id", "io-addr"+(_addr));
    $(_id).appendChild(hr);
    return this; 

}

function t (text) {
    const h3 = ce("h3")
    h3.classList.add("io-title");
    h3.innerText = text;
    $(_id).appendChild(h3);
    return this;
}

function p () {
    var text = "";
    for (var i=0; i<arguments.length; i++) {
	text += (arguments[i] + " ");
    }
    var paras = paragraphs(text);
    //console.log(paras);
    paras.forEach(function(t) {
        const p = ce("p");
	p.innerHTML = t;
        $(_id).appendChild(p);
    });
    return this;
}

function ps () { 
    var text = []
    for (var i=0; i<arguments.length; i++) {
	arguments[i].forEach(function(para) { 
	    p(para);
	});
    }
}

function p_class (cl) {
    var text = "";
    for (var i=1; i<arguments.length; i++) {
	text += (arguments[i] + " ");
    }
    var paras = paragraphs(text);
    //console.log(paras);
    paras.forEach(function(t) {
        const p = ce("p")
	p.classList.add(cl);
        p.innerText = t;
        $(_id).appendChild(p);
    });
    return this;
}


function img (src,style) {
    if (style) {
        const img = ce("img");
	img.setAttribute("src", src);
        img.setAttribute("style", style);
        $(_id).appendChild(img);
    } else { 
        const img = ce("img");
	img.setAttribute("src", src);
        $(_id).appendChild(img);
    }
    return this;
}

function html (h) { 

    const div = ce("div");
    div.innerHTML = h;
    $(_id).appendChild(div);
    return this;
}


function log () {
    if (DEBUG_FLAG) { 
	var text = "";
	for (var i=0; i<arguments.length; i++) {
	    text += (arguments[i] + " ");
	}
	var paras = paragraphs(text);
	//console.log(paras);

	paras.forEach(function(t) {
            const p = ce("p");
            p.classList.add("io-log");
            p.innerText = t;
            $(_id).appendChild(p);
	});
    }
    return this;
}


function space () {
    const div = ce("div");
    div.classList.add("io-hrspace");
    div.innerHTML = "<hr>";
    $(_id).appendChild(div);
    return this;
}


function choices () { 

    function D () { 
	this._options = [];
    }

    D.prototype.group = function (name,options) { 
	if (options.length > 0) { 
	    this._options.push({type:"group",name:name,options:options.map(function(x) { return {text:x[0], run:x[1]}; })});
	}
	return this;
    }

    D.prototype.option = function(text,run) {
	this._options.push({type:"option",text:text,run:run});
	return this;
    }

    D.prototype.input = function(text,run) { 
	this._options.push({type:"input",text:text,run:run});
	return this;
    }

    D.prototype.optionIf = function(cond,text,run) {
	if (cond) { 
	    this.option(text,run)
	}
	return this;
    }

    function clearAndGo (run,arg) { 
	//$(".active-choice").removeClass("io-active-choice");
	document.querySelectorAll(".active-choice,.to-remove").forEach(elt => elt.remove());
	if (arg) { 
	    run(arg);
	} else { 
	    run();
	}
    }

    D.prototype.show = function() { 
	var that = this;
	var bq = ce("ul");
        bq.style.listStyle = "none";
        bq.style.paddingLeft = 0;
        $(_id).appendChild(bq);
	this._options.forEach(function(opt) {
	    var li = ce("li");
            bq.appendChild(li);
	    var opts;
	    if (opt.type==="input") { 
		var id = fresh_id();
                const span = ce("span");
		span.classList.add("io-active-choice");
                span.innerText = opt.text;
                span.addEventListener('click', function() { 
		    if (span.classList.contains("io-active-choice")) {
                        span.classList.add("io-selected-choice");
                        span.classList.remove("io-active-choice");
			clearAndGo(opt.run, $("#"+id).value);
                    }
                });
                li.appendChild(span);
                const span2 = ce("span");
                span2.classList.add("to-remove");
                span2.innerHTML = '<input id="' + id + '" type="text" style="margin-left: 20px; border-color: wheat;">';
                li.appendChild(span2);
		return;
	    }
	    if (opt.type==="group") {
                const span = ce("span");
                span.classList.add("to-remove");
                span.innerText = opt.name+"&nbsp;[&nbsp;";
                li.appendChild(span);
		opts = opt.options;
		var group_name = opt.name;
	    } else {
		opts = [ opt ];
		var group_name = "";
	    }
	    opts.forEach(function(opt,i) {
		if (i > 0) {
                    const span = ce("span");
                    span.classList.add("to-remove");
                    span.innerText = opt.name+"&nbsp;[&nbsp;";
                    li.appendChild(span);
		}
		if (opt.run) { 
                    const span = ce("span");
		    span.classList.add("io-active-choice");
                    span.innerHTML = "<span><span class=\"show-if-selected\">" + group_name + "</span> " + opt.text + "</span>";
                    span.addEventListener('click', function() { 
		        if (span.classList.contains("io-active-choice")) {
                            span.classList.add("io-selected-choice");
                            span.classList.remove("io-active-choice");
			    clearAndGo(opt.run);
                        }
                    });
                    li.appendChild(span);
		} else {
                    const span = ce("span");
                    span.innerHTML = "<span><span class=\"show-if-selected\">" + group_name + "</span> " + opt.text + "</span>";
                    li.appendChild(span);
		}
	    });
	    if (opt.type==="group") {
                const span = ce("span");
                span.classList.add("to-remove");
                span.innerText = "&nbsp;]";
                li.appendChild(span);
	    }
	});

	if (!CLEAR_FLAG) { 
	    scroll(_addr);
	} else { 
	    window.scrollTo(0, 0);   // $("html,body").scrollTop(0);
	}
	return this;
    }

    return new D();
}

const io = {}
io.config = config;
io.splash = splash;
io.img = img;
io.t = t;
io.p = p;
io.ps = ps;
io.p_class = p_class;
io.html = html;
io.choices = choices;
io.newp = newp;
io.space = space;
io.log = log;



function run(game, content)  {
    let config = game.config;
    let title = game.title;
    let subtitle = game.subtitle;
    let author = game.author;
    let state= game.global;
    let passage = game.init;
    let closed_content = content();
    io.config(config);
    io.splash(title, subtitle, author);
    goPassage(state, closed_content, passage, false);
}

function goPassage (state, content, key, clear) { 
    console.log('Passage:', key);
    if (clear) { 
	io.newp();
    }
    if (!(key in content)) {
	io.html('<span style="color: red;"><b>ERROR: No passage ' + key + '</b></span>');
    }
    else { 
	content[key](state);
    }
}

const engine = {}
engine.goPassage = goPassage;
engine.run = run;
    `)
}
