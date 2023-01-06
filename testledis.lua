request = function()
    param_value1 = math.random(1,100)
    param_value2 = math.random(1,100)
    path = '/?p1='..param_value1..'&p2='..param_value1..'&pais=7'
    return wrk.format("GET", path)
end