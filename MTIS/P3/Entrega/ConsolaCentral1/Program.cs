// Nikita Polyanskiy Y4441167L
using System.Text;
using Apache.NMS;
using Apache.NMS.ActiveMQ;

class Program
{
    private const string brokerUri = "tcp://localhost:61616";

    private const string topicLi1 = "LectIlum1";
    private const string topicLt1 = "LectTemp1";
    private const string topicPi1 = "ActIlum1";
    private const string topicPt1 = "ActTemp1";

    private const string topicLi2 = "LectIlum2";
    private const string topicLt2 = "LectTemp2";
    private const string topicPi2 = "ActIlum2";
    private const string topicPt2 = "ActTemp2";

    static private bool actPi1 = false;
    static private bool actPi2 = false;
    static private bool actPt1 = false;
    static private bool actPt2 = false;

    private const int tempMin = 19;
    private const int tempMax = 25;
    private const int ilumMin = 400;
    private const int ilumMax = 500;

    private const int targetTemp = 22;
    private const int targetIlum = 450;

    static void MyListener(string top)
    {
        IConnectionFactory factory = new ConnectionFactory(brokerUri);
        IConnection connection = factory.CreateConnection();
        ISession session = connection.CreateSession();
        ITopic topic = session.GetTopic(top);
        IMessageConsumer consumer = session.CreateConsumer(topic);
        connection.Start();

        Console.WriteLine("Starting listener - topic:" + top);
        while (true)
        {
            IMessage message = consumer.Receive();

            if (message != null)
            {
                //Console.WriteLine("Received message with ID: " + message.NMSMessageId);
                if (message is ITextMessage textMessage)
                {
                    string messageText = textMessage.Text;
                    Console.WriteLine($"Received text message: {messageText}");
                }
                else if (message is IBytesMessage bytesMessage)
                {
                    byte[] messageBytes = new byte[bytesMessage.BodyLength];
                    bytesMessage.ReadBytes(messageBytes);
                    string messageText = Encoding.UTF8.GetString(messageBytes);
                    //Console.WriteLine($"Received message: {messageText}");
                    AnalyzeMessage(top, messageText);
                }
                else
                {
                    Console.WriteLine(message.GetType().FullName);
                }
            }
        }
    }

    static void AnalyzeMessage(string top, string message)
    {
        string t="";
        string msg="";
        bool sending = false;

        switch (top)
        {
            case topicLi1:
                Console.WriteLine("Ilum Office-1: " + message+" lumens");
                if (actPi1)
                {
                    if (Int32.Parse(message) > ilumMin && Int32.Parse(message) < ilumMax)
                    {
                        Console.WriteLine("Deactivate ilum activator for Office 1");

                        sending = true;
                        t = topicPi1;
                        msg = "ilum:" + "-1";
                        actPi1 = false;
                    }

                }
                else
                {
                    if (Int32.Parse(message) < ilumMin || Int32.Parse(message) > ilumMax)
                    {
                        Console.WriteLine("Start ilum activator for Office 1");
                        sending = true;
                        t = topicPi1;
                        msg = "ilum:" + targetIlum.ToString();
                        actPi1 = true;
                    }
                }

                break;
            case topicLi2:
                Console.WriteLine("Ilum Office-2: " + message + " lumens");
                if (actPi2)
                {
                    if (Int32.Parse(message) > ilumMin && Int32.Parse(message) < ilumMax)
                    {
                        Console.WriteLine("Deactivate ilum activator for Office 2");

                        sending = true;
                        t = topicPi2;
                        msg = "ilum:" + "-1";
                        actPi2 = false;
                    }

                }
                else
                {
                    if (Int32.Parse(message) < ilumMin || Int32.Parse(message) > ilumMax)
                    {
                        Console.WriteLine("Start ilum activator for Office 2");
                        sending = true;
                        t = topicPi2;
                        msg = "ilum:" + targetIlum.ToString();
                        actPi2 = true;
                    }
                }
                break;
            case topicLt1:
                Console.WriteLine("Temp Office-1: " + message + "ºC");
                if (actPt1)
                {
                    if (Int32.Parse(message) > tempMin && Int32.Parse(message) < tempMax)
                    {
                        Console.WriteLine("Deactivate temp activator for Office 1");

                        sending = true;
                        t = topicPt1;
                        msg = "temp:" + "-1";
                        actPt1 = false;
                    }

                }
                else
                {
                    if (Int32.Parse(message) < tempMin || Int32.Parse(message) > tempMax)
                    {
                        Console.WriteLine("Start temp activator for Office 1");
                        sending = true;
                        t = topicPt1;
                        msg = "temp:" + targetTemp.ToString();
                        actPt1 = true;
                    }
                }
                break;
            case topicLt2:
                Console.WriteLine("Temp Office-2: " + message + "ºC");
                if (actPt2)
                {
                    if (Int32.Parse(message) > tempMin && Int32.Parse(message) < tempMax)
                    {
                        Console.WriteLine("Deactivate temp activator for Office 2");

                        sending = true;
                        t = topicPt2;
                        msg = "temp:" + "-1";
                        actPt2 = false;
                    }

                }
                else
                {
                    if (Int32.Parse(message) < tempMin || Int32.Parse(message) > tempMax)
                    {
                        Console.WriteLine("Start temp activator for Office 2");
                        sending = true;
                        t = topicPt2;
                        msg = "temp:" + targetTemp.ToString();
                        actPt2 = true;
                    }
                }
                break;
            default:
                Console.WriteLine("Error: No topic at AnalyzeMessage()");
                break;
        }

        if (sending)
        {
            send(t,msg);
        }
    }

    static void send(string top,string messageText)
    {
        var connectionFactory = new ConnectionFactory(brokerUri);

        using IConnection connection = connectionFactory.CreateConnection();
        connection.Start();

        using ISession session = connection.CreateSession(AcknowledgementMode.AutoAcknowledge);

        IDestination destination = session.GetTopic(top);

        using IMessageProducer producer = session.CreateProducer(destination);

        ITextMessage message = session.CreateTextMessage(messageText);

        producer.Send(message);

        connection.Close();
    }
    static void Main(string[] args)
    {
        Console.WriteLine("Starting up Central Console...");

        Thread t1 = new Thread(()=> MyListener(topicLi1));
        Thread t2 = new Thread(() => MyListener(topicLi2));
        Thread t3 = new Thread(() => MyListener(topicLt1));
        Thread t4 = new Thread(() => MyListener(topicLt2));

        t1.Start();
        t2.Start();
        t3.Start();
        t4.Start();
    }
}
