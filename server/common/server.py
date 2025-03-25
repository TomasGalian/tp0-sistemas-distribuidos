import socket
import logging
import signal

from common.utils import Bet, store_bets

fields = ['Agencia', 'Nombre', 'Apellido', 'Documento', 'Fecha de nacimiento', 'Numero']

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._client_sock = None

    def signal_handler(self, sig, frame):
        logging.info('action: shutdown | result: in_progress')
        self._server_socket.close()

        if self._client_sock:
            self._client_sock.close()

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

        while True:
            self._client_sock = self.__accept_new_connection()
            self.__handle_client_connection(self._client_sock)
            self._client_sock = None

    def _receive_data(self, client_sock): 
        data = {}

        # Lo ideal es tener un contador para los campos
        for field in fields:
            # Leer un byte que indica la longitud del siguiente campo
            len_field = client_sock.recv(1)
            
            if not len_field:  # Verificar si no hay más datos
                break

            # Convertir la longitud del campo (es un byte)
            len_field = ord(len_field)

            # Leer el campo con la longitud determinada
            buffer = client_sock.recv(len_field)

            # Verificar si recibimos la cantidad correcta de bytes, en caso de que lleguen fragmentados
            while len(buffer) < len_field:
                buffer += client_sock.recv(len_field - len(buffer))

            # Almacenar el campo recibido
            data[field] = buffer.decode('utf-8')
        
        return data
    
    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            bet_fields = self._receive_data(client_sock)

            bet = Bet(bet_fields[fields[0]], bet_fields[fields[1]], bet_fields[fields[2]], bet_fields[fields[3]], bet_fields[fields[4]], bet_fields[fields[5]])
            store_bets([bet])
            
            logging.info(f'action: apuesta_almacenada | result: success | dni: {bet_fields["Documento"]} | numero: {bet_fields["Numero"]}')

            # TODO: Modify the send to avoid short-writes
            # client_sock.send("{}\n".format(bet).encode('utf-8'))
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
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
