request = function()
    path = "/?c=1&p={'O':[1,1,1],'D':0,'C':[1,2,3,4],'F':[[1],[2],[3],[2,4,5]],'E':[1,2,3,4]}"
    return wrk.format("GET", path)
end