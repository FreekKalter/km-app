import xlsxwriter
import json
import httplib

# get data and decodo json
conn = httplib.HTTPConnection("localhost", 4001)
conn.request("GET", "/overview/tijden/2014/2")
resp = conn.getresponse()
if (resp.status != 200):
    exit("request failed")
d= resp.read()

data = json.loads(d)
#print data

# make work book and define some styles for formatting
workbook = xlsxwriter.Workbook('hello.xlsx')
worksheet = workbook.add_worksheet()
bold = workbook.add_format({'bold':True})
grayBg = workbook.add_format({'bg_color': 'gray'})

lightGray = '#707070'
header = bold
header.set_bg_color(lightGray)
header.set_bottom(1)

firstColumn = bold
firstColumn.set_bg_color(lightGray)
firstColumn.set_right(1)



for y,row in enumerate(data):
    for x,item in enumerate(row):
        if (x ==0):
            worksheet.write(y,x, y, firstColumn)
        if (y == 0):
            worksheet.write(y,x, item, header)
        worksheet.write(y+1,x+1, row[item])
workbook.close()
