request = function()
    param_value = math.random(1,49)
    path = '/?c='..param_value..'&p={"O":[1,1,1],"D":0,"C":[1,2,3,4,5,6,7,8,9,10]}'
    return wrk.format("GET", path)
end