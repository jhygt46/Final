function forms(that){
    
    var inputs = new Array();
    var selects = new Array();
    var textareas = new Array();
    var data = new FormData();
    var require = "";
    var func = "";
    var send = true;
    
    $(that).parents('form').find('input').each(function(){
        
        if($(this).attr('require')){
            require = $(this).attr('require').split(" ");
            for(var i=0; i<require.length; i++){

                func = require[i].split("-");
                if(func[0] == "email"){
                    if(!email($(this).val())){
                        send = false;
                        $(this).parent('label').find('.mensaje').html("No es un correo electronico");
                    }else{
                        $(this).parent('label').find('.mensaje').html("");
                    }
                }
                if(func[0] == "distnada"){
                    if(!distnada($(this).val())){
                        send = false;
                        $(this).parent('label').find('.mensaje').html("Debe completar este campo");
                    }else{
                        $(this).parent('label').find('.mensaje').html("");
                    }
                }
                if(func[0] == "distzero"){
                    if(!distzero($(this).val())){
                        send = false;
                        $(this).parent('label').find('.mensaje').html("Debe seleccionar una opcion");
                    }else{
                        $(this).parent('label').find('.mensaje').html("");
                    }
                }
                if(func[0] == "textma"){
                    if(!textma($(this).val(), func[1])){
                        send = false;
                        $(this).parent('label').find('.mensaje').html("Debe tener a lo menos "+func[1]+" caracteres");
                    }else{
                        $(this).parent('label').find('.mensaje').html("");
                    }
                }
                if(func[0] == "textme"){
                    if(!textme($(this).val(), func[1])){
                        send = false;
                        $(this).parent('label').find('.mensaje').html("Debe tener a lo mas "+func[1]+" caracteres");
                    }else{
                        $(this).parent('label').find('.mensaje').html("");
                    }
                }
            }
        }
        
        if($(this).attr('type') == "password"){
            data.append($(this).attr('id'), $(this).val());
            //inputs.push($(this));
        }
        if($(this).attr('type') == "text"){
            data.append($(this).attr('id'), $(this).val());
            //inputs.push($(this));
        }
        if($(this).attr('type') == "number"){
            data.append($(this).attr('id'), $(this).val());
            //inputs.push($(this));
        }
        if($(this).attr('type') == "date"){
            data.append($(this).attr('id'), $(this).val());
            //inputs.push($(this));
        }
        if($(this).attr('type') == "hidden"){
            data.append($(this).attr('id'), $(this).val());
            //inputs.push($(this));
        }
        if($(this).attr('type') == "checkbox" && $(this).is(':checked')){
            data.append($(this).attr('id'), "1");
            //inputs.push($(this));
        }
        if($(this).attr('type') == "checkbox" && !$(this).is(':checked')){
            data.append($(this).attr('id'), "0");
        }
        if($(this).attr('type') == "radio" && $(this).is(':checked')){
            data.append($(this).attr('id'), $(this).val());
            //inputs.push($(this));
        }
        if($(this).attr('type') == "file"){
            var inputFileImage = document.getElementById($(this).attr('id'));
            for(var i=0; i<inputFileImage.files.length; i++){
                var file = inputFileImage.files[i];
                data.append($(this).attr('id'), file);
            }
        }
    });
    $(that).parents('form').find('select').each(function(){
        data.append($(this).attr('id'), $(this).val());
        //selects.push($(this));
    });
    $(that).parents('form').find('textarea').each(function(){
        data.append($(this).attr('id'), $(this).val());
        //textareas.push($(this));
    });
    
    if(send){
        $('.loading').show();
        $.ajax({
            url: "ajax/",
            type: "POST",
            contentType: false,
            data: data,
            dataType: 'json',
            processData: false,
            cache: false,
            success: function(data){
                console.log(data);
                if(data != null){
                    if(data.reload == 1)
                        navlink('pages/'+data.page);
                    if(data.op != null)
                        mensaje(data.op, data.mensaje);
                }
            },
            error: function(){}
        });
    }
    return false;
    
}
function mensaje(op, mens){
    
    if(op == 1){
        var type = "success";
        var timer = 3000;
    }
    if(op == 2){
        var type = "error";
        var timer = 6000;
    }
    if(op == 3){
        var type = "warning";
        var timer = 6000;
    }
    swal({
        title: "",
        text: mens,
        html: true,
        timer: timer,
        type: type
    });
    
}
function deleteinfo(inputs, selects, textareas){
    
    for(var i=0; i<inputs.length; i++){
        if(inputs[i].attr('type') == "text" || inputs[i].attr('type') == "date"){
            inputs[i].val("");
        }
        if(inputs[i].attr('type') == "hidden" && inputs[i].attr('id') != "accion"){
            inputs[i].val(0);
        }
        if(inputs[i].attr('type') == "checkbox"){
            inputs[i].prop('checked', false);
        }
    }
    
    for(var i=0; i<selects.length; i++){
        if(!selects[i].hasClass("nodelete"))
            selects[i].val(0);
    }
    for(var i=0; i<textareas.length; i++){
        textareas[i].html("");
    }
    
}
function distzero(x){
    if(x != 0){
        return true;
    }
    return false;
}
function distnada(x){
    if(x != ""){
        return true;
    }
    return false;
}
function email(x){
    var expr = /^([a-zA-Z0-9_\.\-])+\@(([a-zA-Z0-9\-])+\.)+([a-zA-Z0-9]{2,4})+$/;
    if ( !expr.test(x) )
    return false;
    return true;
}
function textma(x, i){
    if(x.length > i){
        return true;
    }
    return false;
}
function textme(x, i){
    if(x.length < i){
        return true;
    }
    return false;
}
function eliminar(accion, id, tipo, name){
    var msg = {
        title: "Eliminar "+tipo, 
        text: "Esta seguro que desea eliminar a "+name, 
        confirm: "Si, deseo eliminarlo",
        name: name,
        accion: accion,
        id: id,
    };
    confirm(msg);
}
function add_ses(id, correo){
    var msg = {
        title: "Correo agregado a SES", 
        text: "Desea cambiar el estado del correo "+correo, 
        confirm: "Si, deseo cambiarlo",
        name: correo,
        accion: 'add_ses',
        id: id,
    };
    confirm_warning(msg);   
}
function add_dns(id, dominio){
    var msg = {
        title: "Dns agregada en Google", 
        text: "Desea cambiar el estado del dominio "+dominio, 
        confirm: "Si, deseo cambiarlo",
        name: dominio,
        accion: 'add_dns',
        id: id,
    };
    confirm_warning(msg);
}
function add_ssl(id, dominio){
    var msg = {
        title: "Solicitud SSL", 
        text: "Desea cambiar el estado del dominio "+dominio, 
        confirm: "Si, deseo cambiarlo",
        name: dominio,
        accion: 'add_ssl',
        id: id,
    };
    confirm_warning(msg);
}
function confirm(message){
    
    swal({   
        title: message['title'],   
        text: message['text'],   
        type: "error",   
        showCancelButton: true,   
        confirmButtonColor: "#DD6B55",   
        confirmButtonText: message['confirm'],   
        closeOnConfirm: false,
        showLoaderOnConfirm: true
    }, function(isConfirm){

        if(isConfirm){
            
            var send = {accion: message['accion'], id: message['id'], nombre: message['name']};
            $.ajax({
                url: "ajax/",
                type: "POST",
                data: send,
                success: function(data){
                    
                    var timer = (data.timer === undefined) ? 2000 : data.timer ;

                    setTimeout(function(){  
                        swal({
                            title: data.titulo,
                            text: data.texto,
                            type: data.tipo,
                            timer: timer,
                            showConfirmButton: false
                        });
                        if(data.reload)
                            navlink('pages/'+data.page);
                    }, 10);

                }, error: function(){}
            });

        }
    });
    
}
function confirm_warning(message){
    
    swal({   
        title: message['title'],   
        text: message['text'],   
        type: "warning",   
        showCancelButton: true,   
        confirmButtonColor: "#DD6B55",   
        confirmButtonText: message['confirm'],   
        closeOnConfirm: false,
        showLoaderOnConfirm: true
    }, function(isConfirm){

        if(isConfirm){
            
            var send = {accion: message['accion'], id: message['id'], nombre: message['name']};
            
            $.ajax({
                url: "ajax/",
                type: "POST",
                data: send,
                success: function(data){
                    
                    var timer = (data.timer === undefined) ? 2000 : data.timer ;

                    setTimeout(function(){  
                        swal({
                            title: data.titulo,
                            text: data.texto,
                            type: data.tipo,
                            timer: timer,
                            showConfirmButton: false
                        });
                        if(data.reload)
                            navlink('pages/'+data.page);
                    }, 10);

                }, error: function(){}
            });
 
        }
        
    });
    
}