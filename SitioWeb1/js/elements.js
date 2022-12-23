function getinputval(val, datos){
    if(datos !== null){
        if(datos[val] !== undefined){
            return datos[val];
        }else{
            return '';
        }
    }else{
        if(val.includes('-')){
            var name = val.split('-');
            return getlocaljson(name[0], {})[name[1]] || '' ;
        }
    }
}
function setvalue(){
    var name = this.getAttribute('name');
    if(name.includes('-')){
        var data = name.split('-');
        var obj = getlocaljson(data[0], {});
        obj[data[1]] = this.value;
        sl(data[0], obj);
    }
}
function el(obj){
    var el = document.createElement(obj.div || 'div');
    if(obj.onclick !== undefined){ el.onclick = obj.onclick }
    if(obj.clase){ el.className = obj.clase; }
    if(obj.value){ el.innerHTML = obj.value; }
    for(key in obj.attr){
        if(obj.attr.hasOwnProperty(key)){
            el.setAttribute(key, obj.attr[key])
        }
    }
    return el;
}
function input(obj, datos = null){

    var el = document.createElement(obj.input || 'input');
    if(obj.onclick !== undefined){ el.onclick = obj.onclick }
    if(obj.onkeyup !== undefined){ el.onkeyup = obj.onkeyup }

    if(obj.attr !== undefined){
        for(key in obj.attr){
            if(obj.attr.hasOwnProperty(key)){
                el.setAttribute(key, obj.attr[key])
            }
        }
        if(obj.attr.name !== undefined){ 
            el.setAttribute('value', getinputval(obj.attr.name, datos))
        }
    }
    if(obj.opt !== undefined){
        var i = 0;
        for(op of obj.opt){
            var opt = document.createElement("option");
            if(op.value !== undefined && op.text !== undefined){
                opt.value = op.value;
                opt.text = op.text;
            }else{
                opt.value = i;
                opt.text = op;
                i++;
            }
            el.add(opt, null);
        }
    }
    return el;
}