import socket
import logging
import signal

from common.utils import Bet, store_bets

fields = ['Agencia', 'Nombre', 'Apellido', 'Documento', 'Fecha de nacimiento', 'Numero']

class LotteryCenter:
    def __init__(self, port, listen_backlog):
        # Initialize lottery socket
        self._lottery_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._lottery_socket.bind(('', port))
        self._lottery_socket.listen(listen_backlog)
        self._client_sock = None

    def handler_Sigterm(self, sig, frame):
        logging.debug('action: shutdown | result: in_progress')
        if self._client_sock:
            self._client_sock.close()

        self._lottery_socket.close()
        logging.info('action: shutdown | result: success')
        exit(0)

    def run(self):
        """
        Run the Lottery Center
        Loterry Center accept new connections and establishes a
        communication with an agency. After agency with communucation
        finishes, servers starts to accept new connections again
        """

        # Register signal handler
        signal.signal(signal.SIGTERM, self.handler_Sigterm)

        # Main loop
        while True:
            self._client_sock = self.__accept_new_connection()
            logging.debug('action: accept_connections | result: success')
            self.__handle_client_connection(self._client_sock)
            self._client_sock = None

    def _receive_bet(self, client_sock): 
        data = {}

        for field in fields:
            len_field = client_sock.recv(1)
            
            if not len_field:  
                break

            len_field = ord(len_field)

            buffer = client_sock.recv(len_field)

            while len(buffer) < len_field:
                buffer += client_sock.recv(len_field - len(buffer))

            data[field] = buffer.decode('utf-8')
                                
        bet = Bet(*(data[field] for field in fields))
        return bet
        
    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            logging.debug('action: receive_action | result: in_progress')
            action = client_sock.recv(1)
            if not action:
                client_sock.close()
                return
            
            action = int.from_bytes(action, byteorder='big')
            
            if action == 1:
                bet = self._receive_bet(client_sock)
                store_bets([bet])
                logging.info(f'action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}')
                
            logging.debug('action: send_ack | result: in_progress')
            client_sock.send(b'\x11')
        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._lottery_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
