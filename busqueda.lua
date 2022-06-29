request = function()
    param_value = math.random(1,500)
    path = '/?c='..param_value..'&p={"O":[1,1,1],"D":0,"C":[10,20,30,40,50,60,70,80]}'
    return wrk.format("GET", path)
end