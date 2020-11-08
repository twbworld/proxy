#!/bin/bash

#分享链接
proxyList=(
  'trojan://trojan@tg.twbhub.cf:443?sni=tg.twbhub.cf'
  'vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIiIsDQogICJhZGQiOiAiNDUuNzYuMTk0Ljc5IiwNCiAgInBvcnQiOiAiMjA4MyIsDQogICJpZCI6ICJlNjYwOTliMC00ZmY2LTQwMGEtYmE5ZC1lZDMyOTYyNTQ5ZGMiLA0KICAiYWlkIjogIjAiLA0KICAibmV0IjogInRjcCIsDQogICJ0eXBlIjogIm5vbmUiLA0KICAiaG9zdCI6ICIiLA0KICAicGF0aCI6ICIiLA0KICAidGxzIjogIiINCn0='
)
#订阅文件
subscriptionPath='../blog/static/proxy.html'

function createFile () {
  #拼接数组
  proxyStr=''
  for i in ${proxyList[@]};do
    if [ "$proxyStr" = '' ]
    then
      proxyStr=$proxyStr$i;
    else
      proxyStr=`echo -e "${proxyStr}\n${i}"`;
    fi
  done

  #加密
  proxyStrBase64=$(base64 <<< $proxyStr)

  touch $subscriptionPath
  cat > $subscriptionPath <<EOF
${proxyStrBase64}
EOF
  echo "订阅文件生成"
}

if [ -f $subscriptionPath ];then
rm -rf $subscriptionPath
echo "订阅文件已删除"
fi
createFile
