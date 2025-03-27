import socket
import logging
import signal

from common.utils import Bet, store_bets, has_won, load_bets

fields = ['Agencia', 'Nombre', 'Apellido', 'Documento', 'Fecha de nacimiento', 'Numero']

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._client_sockets = {}

    def signal_handler(self, sig, frame):
        logging.info('action: shutdown | result: in_progress')
        
        for _client_sock in self._client_sockets.values():
            if _client_sock:
                _client_sock.close()

        self._server_socket.close()
        logging.info('action: shutdown | result: success')
        exit(0)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        signal.signal(signal.SIGTERM, self.signal_handler)

        for i in range(0, 5): # Total agencys
            self._client_sock = self.__accept_new_connection()
            self.__handle_client_connection(self._client_sock)

        # Hacer el sorteo
        logging.info('action: sorteo | result: success')
        bets_list = load_bets()
        for agency_id, _client_sock in self._client_sockets.items():
            if ord(_client_sock.recv(1)) == 255:  # Receive the signal to start the draw
                for bet in bets_list:
                    if has_won(bet) and bet.agency == agency_id:
                        _client_sock.send(len(bet.document).encode('utf-8')) # TODO avoid short writes
                        _client_sock.send(bet.document.encode('utf-8')) # TODO avoid short writes
                        
                _client_sock.send(b'0')  # Send a byte to signal the end of the draw

    def _receive_data(self, client_sock): 
        bets = []

        total_bets = client_sock.recv(1)
        if ord(total_bets) == 0:  
            # Flush the socket to ensure no leftover data remains (TODO verify)
            client_sock.recv(1024)
            return None

        total_bets = ord(total_bets)

        for _ in range(total_bets):
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
            
            bets.append(data)
        
        return bets
    
    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            while True:  # Keep processing chunks until the client closes the connection
                bets_fields = self._receive_data(client_sock)
                if not bets_fields:  
                    break

                addr = client_sock.getpeername()
                self._client_sockets[bets_fields[0]['Agencia']] = client_sock

                for bet_fields in bets_fields:  
                    bet = Bet(*(bet_fields[field] for field in fields))
                    store_bets([bet])
                
                logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets_fields)}')

                ack = b'\x01'  # Send a byte as ACK (confirmation)
                client_sock.send(ack)
        except OSError as e:
            logging.info(f'action: apuesta_recibida | result: fail | cantidad: {len(bets_fields)}')
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
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
