request = function()
    param_value1 = math.random(140910,141010)
    param_value2 = math.random(140910,141010)
    param_value3 = math.random(140910,141010)
    param_value4 = math.random(140910,141010)
    path = '/autoCuad?c=['..param_value1..','..param_value2..','..param_value3..','..param_value4..']&u=12'
    return wrk.format("GET", path)
end