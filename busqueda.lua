request = function()
    param_value = math.random(1,49)
    path = '/?c='..param_value..'&p={"O":[1,1,1],"D":0,"C":[100,200,300,400]}'
    return wrk.format("GET", path)
end