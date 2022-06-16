request = function()
    param_value = math.random(1,49)
    path = '/?c='..param_value..'&p={"O":[1,1,1],"D":0,"C":[1,2,3,4,50,60,70,80,90,100]}'
    return wrk.format("GET", path)
end