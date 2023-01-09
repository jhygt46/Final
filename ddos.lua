request = function()
    b1 = math.random(0,255)
    b2 = math.random(0,255)
    b3 = math.random(0,255)
    b4 = math.random(0,255)
    path = '/?ip='..b1..'.'..b2..'.'..b3..'.'..b4..':62534'
    return wrk.format("GET", path)
end