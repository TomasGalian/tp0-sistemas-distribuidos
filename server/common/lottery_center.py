import socket
import logging
import signal
import os

from common.utils import Bet, store_bets, has_won, load_bets

fields = ['Agencia', 'Nombre', 'Apellido', 'Documento', 'Fecha de nacimiento', 'Numero']

class LotteryCenter:
    def __init__(self, port, listen_backlog):
        # Initialize lottery socket
        self._lottery_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._lottery_socket.bind(('', port))
        self._lottery_socket.listen(listen_backlog)
        self._client_sockets = {}

    def handler_Sigterm(self, sig, frame):
        logging.debug('action: shutdown | result: in_progress')
        for _client_sock in self._client_sockets.values():
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
        # Removed unused variable agency_number
        agency_number = os.getenv('AGENCY_NUMBER')

        for i in range(0, int(agency_number)): # Total agencys
            self._client_sock = self.__accept_new_connection()
            logging.debug('action: accept_connections | result: success')
            self.__handle_client_connection(self._client_sock)

        # Hacer el sorteo
        winners = self._get_winners()
        logging.info('action: sorteo | result: success')
        for agency_id, _client_sock in self._client_sockets.items():
            action = _client_sock.recv(1)
            if action == b'\xF0':  # Receive the signal to deliver winners
                winner_agency = winners.get(agency_id, [])
                self._send_winners(_client_sock, winner_agency)

    def _get_winners(self):
        bets_list = load_bets()
        winners = {}
        for bet in bets_list:
            if has_won(bet):
                if bet.agency not in winners:
                    winners[bet.agency] = []
                winners[bet.agency].append(bet)
        return winners

    def _send_winners(self, client_sock, winners):        
        bets_serialized = bytearray()
        # Supose that len(winners) is less than 255
        bets_serialized.extend(len(winners).to_bytes(1, byteorder='big')) 
        
        for winner in winners:
            bet_serialized = self._serialize_bet(winner)
            bets_serialized.extend(bet_serialized)
        
        client_sock.sendall(bets_serialized)

    def _serialize_bet(self, bet):
        serialized_bet = bytearray()

        # Serialize agency_id
        serialized_bet.append(len(str(bet.agency))) 
        serialized_bet.extend(str(bet.agency).encode('utf-8')) 

        serialized_bet.append(len(bet.first_name))  
        serialized_bet.extend(bet.first_name.encode('utf-8'))  

        serialized_bet.append(len(bet.last_name)) 
        serialized_bet.extend(bet.last_name.encode('utf-8'))  

        serialized_bet.append(len(bet.document))  
        serialized_bet.extend(bet.document.encode('utf-8'))  

        serialized_bet.append(len(str(bet.birthdate)))  
        serialized_bet.extend(str(bet.birthdate).encode('utf-8'))  

        serialized_bet.append(len(str(bet.number)))
        serialized_bet.extend(str(bet.number).encode('utf-8'))  

        # Return serialized data as bytes
        return bytes(serialized_bet)

    def _receive_bet(self, client_sock, total_bets):
        bets = []
        bet = {}
        for _ in range(total_bets):
            for field in fields:
                len_field = client_sock.recv(1)
                if not len_field:  
                    break

                len_field = ord(len_field)
                buffer = client_sock.recv(len_field)
                while len(buffer) < len_field:
                    buffer += client_sock.recv(len_field - len(buffer))

                bet[field] = buffer.decode('utf-8')
            
            bets.append(Bet(bet['Agencia'], bet['Nombre'], bet['Apellido'], bet['Documento'], bet['Fecha de nacimiento'], bet['Numero']))
        
        return bets
    
    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            logging.debug('action: receive_action | result: in_progress')
            # Use action field to know how many bets will be received
            action = client_sock.recv(1)
            action = int.from_bytes(action, byteorder='big')
            while action >= 1:
                bets = self._receive_bet(client_sock, action)
                for bet in bets:
                    store_bets([bet])
                logging.info(f'action: apuesta_recibida | result: success | cantidad: {action}')
                    
                logging.debug('action: send_ack | result: in_progress')
                client_sock.send(b'\xFF')

                action = client_sock.recv(1)
                action = int.from_bytes(action, byteorder='big')
            self._client_sockets[bets[0].agency] = self._client_sock
        except OSError as e:
            logging.info(f'action: apuesta_recibida | result: fail | cantidad: {action}')

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
