run:
	docker-compose up --build
	
clear:
	docker-compose down --remove-orphans

test:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.test.yml down --remove-orphans
