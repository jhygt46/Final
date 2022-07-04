request = function()
    param_value1 = math.random(140910,141010)
    param_value2 = math.random(140910,141010)
    path = '/auto?c=['..param_value1..','..param_value2..']'
    return wrk.format("GET", path)
end