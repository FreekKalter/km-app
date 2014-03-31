import xlsxwriter
import json
import httplib
from os.path import join
from xlrd import open_workbook

from xlutils.copy import copy
rb= open_workbook('uren_template.xlsx',   on_demand=True)

wb = copy(rb)

ws = wb.get_sheet(0)
ws.write(2,7, 'hello')

wb.save('output.xlsx')
# get data and decodo json
conn = httplib.HTTPConnection("localhost", 4001)
conn.request("GET", "/overview/tijden/2014/2")
resp = conn.getresponse()
if (resp.status != 200):
    exit("request failed")
d= resp.read()

data = json.loads(d)
#print data

## make work book and define some styles for formatting
workbook = xlsxwriter.Workbook('out.xlsx')
worksheet = workbook.add_worksheet()
#print workbook.worksheets()

#lightGray = '#707070'
#bold = workbook.add_format({'bold':True})
#grayBg = workbook.add_format({'bg_color': 'gray'})

#boldGray = bold
#boldGray.set_bg_color(lightGray)

#header = boldGray
#header.set_bottom(1)
#for x,item in enumerate(data[0]):
    #worksheet.write(0,x+1, item, boldGray)

#firstColumn = boldGray
#firstColumn.set_right(1)
#for y in range(len(data)):
    #worksheet.write(y+1,0, y+1,boldGray)

for y,row in enumerate(data):
    for x,item in enumerate(row):
        worksheet.write(y+1,x+1, row[item])
workbook.close()
