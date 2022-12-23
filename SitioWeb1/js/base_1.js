var ListaBase = [];
for(var i=1; i<=1000; i++){ListaBase.push({i:i,n:makeid(6)});}
var paths = [{d:"M 135 29 C 1 13 23 487 336 458 C 885 168 872 84 135 29", f:"green"}]
var busqueda = { id: 0, Nombre: ""};
function AddSVG(d, f){newpath = document.createElementNS('http://www.w3.org/2000/svg', 'path');newpath.setAttributeNS(null, "d", d);newpath.setAttributeNS(null, "fill", f);GI("s").appendChild(newpath);}
function GI(d){return document.getElementById(d)}
function GC(d,n){return document.getElementsByClassName(d)[n]}
function SetCategoria(id, nombre){ busqueda.id=id; busqueda.nombre=nombre; GC("q1",0).style.display = "none";}
function BackCategoria(){GC("q1",0).style.display = "block";}
function CreateLi(l){var li = document.createElement("li");li.setAttribute("id", l.i);li.appendChild(document.createTextNode(l.n));return li;}
function Busqueda(val){var res = [];for(var m of ListaBase){if (m.n.indexOf(val) == 0){res.push({i:m.i,n:m.n});}}return res;}
function foundAuto(val){for(x of arrSearch){if (x.indexOf(val) == 0){return true;}}return false;}
function SelectPos(items){for (var i=0, ilen=items.length; i<ilen; i++){if (items[i].className == "select"){return i;}}return -1;}
function foundAutoExacto(val){for(x of arrSearch){if(x == val){return true;}}return false;}

function myScript(){
    paths.forEach(function(val){AddSVG(val.d, val.f)});
    GI("ai-search").addEventListener('input', inputHandler);
    GI("ai-search").addEventListener('propertychange', inputHandler);
    //https://stackoverflow.com/questions/4019894/get-all-li-elements-in-array
    GI("list-search").addEventListener('click', (event) => {SetCategoria(event.target.getAttribute("id"),event.target.innerHTML);});
    GI("list-search").addEventListener('mouseover', (event) => {var lis = event.target.parentNode.getElementsByTagName("li");for(var i=0, ilen=lis.length; i<ilen; i++){lis[i].className = "";}event.target.className = "select";});
    GI("list-search").addEventListener('mouseout', (event) => {var lis = event.target.parentNode.getElementsByTagName("li");for(var i=0, ilen=lis.length; i<ilen; i++){lis[i].className = "";}});
}

function GetNumValue(value){
    var j = -1;
    for(var i=value.length, ilen=2; i>=ilen; i--){
        if(!foundAutoExacto(value.substring(0, i))){
            arrSearch.push(value.substring(0, i));
            j++;
        }
    }
    return j;
}
function AutoComplete(val1, value){

    isDaemon = false;
    var num = GetNumValue(value);
    var lsearch = GI("list-search");

    console.log("SEARCH: "+value+" - NUM: "+num);

    if (self.fetch) {
        fetch("http://localhost:81/json/"+value+"/"+num)
        .then(response => response.json())
        .then(data => {
            
            //console.log(lsearch, data);

        });
    }else{
        //XMLHttpRequest
    }
}

var arrSearch = [];
var startTime, value, val1;
var isDaemon = false;

function Daemon(){
    setTimeout(function(){
        if((new Date() - startTime)/(value.length - val1.length + 1) > 200){
            AutoComplete(val1, value);
        }else{
            Daemon();
        }
    },10);
}

const inputHandler = function(e) {
    var lsearch = GI("list-search");
    value = e.target.value;
    if (value != "") {
        var local = Busqueda(value);
        lsearch.innerHTML = "";
        for(var x of local){
            lsearch.appendChild(CreateLi(x));
        }
        if(!isDaemon && local.length < 7 && !foundAuto(value)){
            Daemon();
            isDaemon = true;
            val1 = value;
            startTime = new Date();
        }
        lsearch.style.display = "block";
    }else{
        lsearch.style.display = "none";
    }
}

window.addEventListener('keydown', function (e) {
    var lsearch = GI("list-search");
    var islsearch = (lsearch.style.display == "block") ? true : false ;
    var items = lsearch.getElementsByTagName("li");
    if (islsearch) {
        var pos = SelectPos(items);
        switch(e.key) {
            case "ArrowDown":
                if(pos > -1 && pos < items.length - 1){items[pos].className = "";}
                if(pos < items.length - 1){items[pos+1].className = "select";}
                break;
            case "ArrowUp":
                if(pos > 0){items[pos].className = "";items[pos-1].className = "select";}
                break;
            case "Enter":
                SetCategoria(items[pos].getAttribute("id"),items[pos].innerHTML);
                break;
        }
    }
}, false);

// TEMP //
function makeid(length) {
    var result           = '';
    var characters       = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    var charactersLength = characters.length;
    for ( var i = 0; i < length; i++ ) {
        result += characters.charAt(Math.floor(Math.random() * charactersLength));
    }
    return result;
}
// TEMP //

/*
function Animate(){
    var obj1 = GC("q1",0)
    var obj2 = GC("q2",0)
    console.log("Left: "+obj1.offsetLeft)
    console.log("Left: "+obj2.offsetLeft)
    obj1.style.left = "-50%"
    obj2.style.left = "50%"
    //Animar(obj1,"left",10,100)
    //Animar(obj2,"left",10,100)
}
function Animar(el,style,val,time){
    var m = parseInt(el.style[style],10)
    el.style[style] = m
    setTimeout(function(){
        Animar(el,style,m,time);
    }, time)
}
*/