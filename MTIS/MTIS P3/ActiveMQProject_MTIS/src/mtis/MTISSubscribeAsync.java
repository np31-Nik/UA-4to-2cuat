package mtis;

import javax.jms.*;

import org.apache.activemq.ActiveMQConnection;
import org.apache.activemq.ActiveMQConnectionFactory;
import java.util.concurrent.CountDownLatch;

public class MTISSubscribeAsync  implements MessageListener {

	/**
	 * @param args
	 */
	public static void main(String[] args) {
		
		// URL of the JMS server.
		String url = "tcp://localhost:61616";
		// Name of the topic we will receive messages from
	    String subject = "historiales";
	    CountDownLatch latch = new CountDownLatch(1);
	    try{
			// Getting JMS connection from the server
	        ConnectionFactory connectionFactory
	            = new ActiveMQConnectionFactory(url);
	        Connection connection = connectionFactory.createConnection();
	        connection.start();
	
	        // Creating session for seding messages
	        Session session = connection.createSession(false,
	            Session.AUTO_ACKNOWLEDGE);
	
	        // Getting the topic
	        Destination destination = session.createTopic(subject);
	
	        // MessageConsumer is used for receiving (consuming) messages
	        MessageConsumer consumer = session.createConsumer(destination);
	
	        // Here we receive the message.
	        // By default this call is blocking, which means it will wait
	        // for a message to arrive on the topic.
	        consumer.setMessageListener(new MTISSubscribeAsync());
	        latch.await();
	        consumer.close();
	        connection.close();
	    }catch (Exception e)
	    {
	    	System.out.println("Error: "+e);
	    }

	}

    @Override
    public void onMessage(Message message) {
        try {
            if (message instanceof TextMessage) {
	            TextMessage textMessage = (TextMessage) message;
	            System.out.println("Received message '"
	                + textMessage.getText() + "'");
            }
        } catch (JMSException e) {
            System.out.println("Got a JMS Exception!");
        }
    }
}
