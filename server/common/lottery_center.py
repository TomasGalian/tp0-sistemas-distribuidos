import socket
import logging
import signal
import multiprocessing


from common.utils import Bet, store_bets, has_won, load_bets

fields = ['Agencia', 'Nombre', 'Apellido', 'Documento', 'Fecha de nacimiento', 'Numero']

class LotteryCenter:
    def __init__(self, port, listen_backlog, max_agencies):
        # Initialize lottery socket
        self._lottery_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._lottery_socket.bind(('', port))
        self._lottery_socket.listen(listen_backlog)
        self._client_sockets = {}
        self._max_agencies = int(max_agencies)
        self._manager = multiprocessing.Manager()  # Create a Manager
        self._winners_by_agency = self._manager.dict()  # Shared dictionary for winners

        self._barrier_bet = multiprocessing.Barrier(int(max_agencies) + 1)  # Para sincronizar recepción de apuestas
        self._barrier_sorteo = multiprocessing.Barrier(int(max_agencies) + 1)  # Para sincronizar después del sorteo


        self._lock_file = multiprocessing.Lock()  # Lock para evitar race conditions en el archivos
        self._lock_client_sockets = multiprocessing.Lock()
        self._lock_winners = multiprocessing.Lock()


    def handler_Sigterm(self, sig, frame):
        logging.debug('action: shutdown | result: in_progress')
        with self._lock_client_sockets:
            for _client_sock in self._client_sockets.values():
                if _client_sock:
                    _client_sock.close()

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

        processes = []

        # Recive connections and bets
        for i in range(0, self._max_agencies): # Total agencys
            client_sock = self.__accept_new_connection()
            logging.debug('action: accept_connections | result: success')
            p = multiprocessing.Process(target=self.__handle_client_connection, args=(client_sock, self._lock_file, self._lock_client_sockets ,self._lock_winners , self._barrier_bet, self._barrier_sorteo))
            p.start()
            processes.append(p)


        # Once all bets are received, the lottery center will start the lottery
        self._barrier_bet.wait()
        with self._lock_winners:
            self._winners_by_agency.update(self._get_winners_by_agency())  # Update shared dictionary
        logging.info('action: sorteo | result: success')
        self._barrier_sorteo.wait()

        for p in processes:
            p.join()

    def _get_winners_by_agency(self):
        with self._lock_file:
            bets_list = load_bets()
        winners = {}
        for bet in bets_list:
            if has_won(bet):
                if bet.agency not in winners:
                    winners[bet.agency] = []
                winners[bet.agency].append(bet)
        return winners

    def _handle_winners(self, client_sock, winners):
        for winner in winners:
            serialized_winner = self._serialize_bet(winner)
            client_sock.sendall(serialized_winner)

    def _send_winners(self, client_sock, winners):        
        bets_serialized = bytearray()
        # Supose that len(winners) is less than 255
        bets_serialized.extend(len(winners).to_bytes(1, byteorder='big')) 
        
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
    
    def __handle_client_connection(self, client_sock, lock_file, lock_client_sockets, lock_winners, barrier_bet, barrier_sorteo):
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
                with lock_file:                
                    for bet in bets:
                        store_bets([bet])
                logging.info(f'action: apuesta_recibida | result: success | cantidad: {action}')
                    
                logging.debug('action: send_ack | result: in_progress')
                client_sock.send(b'\xFF')

                action = client_sock.recv(1)
                action = int.from_bytes(action, byteorder='big')
            
            agency_id = bets[0].agency
            with lock_client_sockets:
                self._client_sockets[agency_id] = client_sock

            barrier_bet.wait()
            barrier_sorteo.wait()

            action = client_sock.recv(1)
            if action == b'\xF0':  # Receive the signal to deliver winners
                with lock_winners:
                    winner_agency = self._winners_by_agency.get(agency_id, [])
                
                self._send_winners(client_sock, winner_agency)

        except OSError as e:
            logging.info(f'action: apuesta_recibida | result: fail | cantidad: {action}')
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
