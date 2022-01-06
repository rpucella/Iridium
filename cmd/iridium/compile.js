import os
import json
import sys
import re


def run (directory, target):

    # loop through the files in the directory

    result = {}
    
    with os.scandir(directory) as files:
        for f in files:
            if f.name.endswith('.txt'):
                name = f.name[:len(f.name) - 4]
                print(f" {f.name}")
                with open(os.path.join(directory, f.name), 'rt') as fp:
                    chunks = []
                    current_chunk = ''
                    for line in fp:
                        sline = line.strip()
                        if sline and sline.startswith('# '):
                            # line starting with #+space - comment
                            pass
                        elif sline and sline.startswith('# '):
                            # line starting with # - its own chunk
                            if current_chunk:
                                chunks.append(current_chunk)
                            chunks.append(sline)
                            current_chunk = ''
                        elif sline:
                            # line not starting with # - add to existing chunk
                            current_chunk += sline + ' '
                        else:
                            # not a line, stove current chunk if not empty
                            if current_chunk:
                                chunks.append(current_chunk)
                                current_chunk = ''
                    if current_chunk:
                        chunks.append(current_chunk)
                    classified_chunks = [classify(chunk) for chunk in chunks]
                    body = compile(classified_chunks)
                    code = f"content[{json.dumps(name)}] = (function(state) {{ {body} }});\n"
                    result[name] = code
                    
    ##print(json.dumps(result, indent=2))
    
    content = ''.join(result.values())
    print ('Writing ' + target)
    with open(target, 'wt') as fp:
        #fp.write(f"const engine = require('./engine.js');\nconst io = require('./io.js');\n\nconst content  = function(fn) {{ let content = {{}};\n{content}; return content; }}\n\nmodule.exports = content;\n")
        fp.write(f"const content  = function(fn) {{ let content = {{}};\n{content}; return content; }}\n\n")
                 

def classify (chunk):

    m = re.match(r'^#option\s+([^\s]*)\s+"([^"]+)"\s+$', chunk)
    if m:
        return {
            'type': 'option',
            'passage': m.group(1),
            'text': m.group(2)
        }

    return {
        'type': 'text',
        'text': chunk
    }


options = []



def compile (chunks):

    global options
    
    options = []

    result = parse(chunks)
    if not result:
        print('PARSING ERROR')
        print('Chunks = ', chunks)
        raise Exception('Cannot parse chunks')

    if len(result[1]) > 0:
        raise Exception('Leftover chunks after parsing? ' + result[1])

    return f"var c = io.choices(); {result[0]} c.show(); "



def parse (chunks):
    
    if chunks:

        base = parseBase(chunks)
        if base:
            rest = parse(base[1])
            return (base[0] + rest[0], rest[1]) if rest else None

        # // if chunks rest-if chunks
        # var chunk = chunks[0];
        # let check_if = checkType(chunks,"if");
        # if (check_if) { 
        #     ///console.log("Recognizing if",chunks);
        #     let sub_chunks_1 = parseChunks(check_if[1]);
        #     if (sub_chunks_1) { 
        #       let rest_if = parseRestIf(sub_chunks_1[1]);
        #       if (rest_if) { 
        #           let sub_chunks_2 = parseChunks(rest_if[1]);
        #           if (sub_chunks_2) { 
        #               return ["if ("+chunk.expr+") { " + sub_chunks_1[0] + "} " + rest_if[0] + sub_chunks_2[0],
        #                       sub_chunks_2[1]];
        #           } else { 
        #               return null;
        #           }
        #       } else { 
        #           return null;
        #       }
        #     } else { 
        #       return null;
        #     }
        # }

    return ('', chunks)



def parseBase (chunks):

    current_option = None
    
    if chunks:
        
        chunk = chunks[0]

        if chunk['type'] == 'option':
            
            current_option = f"c = c.option({json.dumps(chunk['text'])}, function() {{ engine.goPassage(state, content, {json.dumps(chunk['passage'])}, true); }}); "
            options.append(current_option)
            return (current_option, chunks[1:])

        elif chunk['type'] == 'text':

            return (f"io.p({json.dumps(chunk['text'])}); ", chunks[1:])


        # case "custom":
        #     current_option = `c = c.option(${JSON.stringify(chunk.text)}, function() { fn.${chunk.fun}(state,content); }, true); `;
        #     options = options + current_option;
        #     return [current_option,
        #           chunks.slice(1)];
        #
        # case "code":
        #     return [chunk.code,
        #           chunks.slice(1)];
        
        # case "image":
        #     return ["io.img("+JSON.stringify(`assets/${chunk.file}`)+","+JSON.stringify(chunk.style)+"); ",
        #           chunks.slice(1)];

        # case "call":
        #     let args = ["state"].concat(chunk.args.map((x) => JSON.stringify(x))).join(",");
        #     return ["fn."+chunk.fn+`(${args}); `,
        #           chunks.slice(1)];

        # case "go":
        #     return ["return engine.goPassage(state,content," + JSON.stringify(chunk.passage) + ",false); ",
        #           chunks.slice(1)];

        # case "include":
        #     return ["content[" + JSON.stringify(chunk.passage) + "](state); ",
        #           chunks.slice(1)];
            
        # }

    return None

    

def clean_text (text):

    ### return text.replace(/^"/g,"<q>").replace(/ "/g," <q>").replace(/"$/g,"</q>").replace(/" /g,"</q> ")
    
    return text



if __name__ == '__main__':

    if len(sys.argv) < 3:
        print("USAGE: compile.py input-dir output-file")
    else:
        run(sys.argv[1], sys.argv[2])
