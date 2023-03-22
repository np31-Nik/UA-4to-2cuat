# Nikita Polyanskiy Y4441167L
import stomp
import random
import time
from threading import Thread

minTemp = 0
maxTemp = 50
minIlum = 200
maxIlum = 1000

topicL1 = "ActIlum2"
topicL2 = "ActTemp2"

topicP1 = "LectIlum2"
topicP2 = "LectTemp2"

temp = random.randint(minTemp,maxTemp)
ilum = random.randint(minIlum,maxIlum)

# Increase/decrease ilum/temp every X seconds
speed = 2


class MyListener(stomp.ConnectionListener):
    def on_error(self, frame):
        print('Received an error "%s"' % frame.body)

    def on_message(self, frame):
        print('Received a message "%s"' % frame.body)
        msg = frame.body.split(":")
        type = msg[0]
        target = int(msg[1])
        if(type=="temp"):
            TempControl.activate(target)
        elif(type=="ilum"):
            IlumControl.activate(target)
        else:
            print("Error: wrong type (temp/ilum) received")

    def start(topic):
        print("Starting listener - topic: "+topic)

        conn = stomp.Connection()
        conn.set_listener('', MyListener())
        conn.connect('admin', 'password', wait=True)

        conn.subscribe(destination='/topic/'+topic, id=1, ack='auto')

        while True:
            pass
        conn.disconnect()

class MyPublisher():
    broker_address = 'localhost'
    broker_port = 61613

    def start(topic):
        topic_name = '/topic/'+topic
        print("Starting publisher - topic:"+topic)

        conn = stomp.Connection()
        conn.connect()

        while(True):
            message = ""

            if(topic==topicP2):
                message = str(temp)
            elif(topic==topicP1):
                message = str(ilum)

            conn.send(body=message, destination=topic_name,headers={'content-type': 'text/plain'})
            #print("Sending message: (Topic: "+topic+" | Message: "+str(message)+")")
            time.sleep(5)

        conn.disconnect()

class TempControl():
    active=False
    mode=0
    target=0

    def printTemp():
        global temp
        while(True):
            if(TempControl.mode==1):
                temp+=1
                if(temp>=TempControl.target):
                    TempControl.active=False
                    TempControl.mode=0
            elif(TempControl.mode==-1):
                temp-=1
                if(temp<=TempControl.target):
                    TempControl.active=False
                    TempControl.mode=0
            print("Current temperature: "+str(temp)+"ºC")
            time.sleep(speed)

    def activate(target):
        if(target==-1):
            TempControl.active=False
            TempControl.mode=0
            print("Deactivating temp control")
        else:
            TempControl.active=True
            TempControl.target=target
            if(temp<target):
                TempControl.mode=1
                print("Increasing temperature to "+str(target)+"ºC")
            elif(temp>target):
                TempControl.mode=-1
                print("Decreasing temperature to "+str(target)+"ºC")
            else:
                TempControl.mode=0
                print("Warning: Target temperature is the same as current temperature")
                TempControl.active=False

class IlumControl():
    active=False
    mode=0
    target=0

    def printIlum():
        global ilum
        while(True):
            if(IlumControl.mode==1):
                ilum+=1
                if(ilum>=IlumControl.target):
                    IlumControl.active=False
                    IlumControl.mode=0
            elif(IlumControl.mode==-1):
                ilum-=1
                if(ilum<=IlumControl.target):
                    IlumControl.active=False
                    IlumControl.mode=0
            print("Current light level: "+str(ilum)+" lumens")
            time.sleep(speed)

    def activate(target):
        if(target==-1):
            IlumControl.active=False
            IlumControl.mode=0
            print("Deactivating ilum control")
        else:
            IlumControl.active=True
            IlumControl.target=target
            if(ilum<target):
                IlumControl.mode=1
                print("Increasing light level to "+str(target)+" lumens")
            elif(ilum>target):
                IlumControl.mode=-1
                print("Decreasing light level to "+str(target)+" lumens")
            else:
                IlumControl.mode=0
                print("Warning: Target light level is the same as current light level")
                IlumControl.active=False

def main():
    print("Starting up Office-2...")

    p1 = Thread(target=MyListener.start,kwargs={"topic":topicL1})
    p2 = Thread(target=MyListener.start,kwargs={"topic":topicL2})
    p3 = Thread(target=IlumControl.printIlum)
    p4 = Thread(target=TempControl.printTemp)
    p5 = Thread(target=MyPublisher.start,kwargs={"topic":topicP1})
    p6 = Thread(target=MyPublisher.start,kwargs={"topic":topicP2})

    p1.start()
    p2.start()
    p3.start()
    p4.start()
    p5.start()
    p6.start()

    p1.join()
    p2.join()
    p3.join()
    p4.join()
    p5.join()
    p6.join()

if __name__ == "__main__":
    main()