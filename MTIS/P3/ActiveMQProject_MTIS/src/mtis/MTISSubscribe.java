package mtis;

import javax.jms.*;

import org.apache.activemq.ActiveMQConnectionFactory;

public class MTISSubscribe {

	/**
	 * @param args
	 */
	public static void main(String[] args) {

		// URL of the JMS server.
		String url = "tcp://localhost:61616";
		// Name of the topic we will receive messages from
	    String subject = "historiales";
	    
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
	        Message message = consumer.receive();
	
	        // There are many types of Message and TextMessage
	        // is just one of them. Producer sent us a TextMessage
	        // so we must cast to it to get access to its .getText()
	        // method.
	        if (message instanceof TextMessage) {
	            TextMessage textMessage = (TextMessage) message;
	            System.out.println("Received message '"
	                + textMessage.getText() + "'");
	        }
	        connection.close();
	    }catch (JMSException e)
	    {
	    	System.out.println("Error: "+e);
	    }
	}

}
