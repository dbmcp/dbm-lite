Integrate the content of all Markdown files located in the `d:\dbm\dbm-lite\.trae\` directory, including the specific file `d:\dbm\dbm-lite\.trae\rules\dbm-lite.md` , and implement the complete project code according to the content requirements. After implementation, execute the following automated workflow:

1. Automatically run the run.bat file in the command line environment to compile and start the service
2. Automatically open a browser to access http://localhost:3000
3. Simulate login to the system with the admin user identity
4. Systematically and automatically click on all functional modules, including but not limited to:
   - All page navigation menus
   - All functional buttons
   - All form CRUD operations
   - All data entry scenarios
   - All interactive elements
5. Collect all error information from the front-end console and back-end services comprehensively
6. Perform targeted fixes based on the collected error information

After each fix is completed, repeat the following process:
1. Check and terminate all processes occupying port 3000 (front-end) and port 8080 (back-end)
2. Delete all old executable files and compilation products
3. Recompile the front-end and back-end code
4. Run run.bat to start the service
5. Repeat the above simulated login and function verification process

Continuously cycle through the above development, testing, and repair processes until the following conditions are met:
1. All page function buttons can be used normally
2. All form operations (CRUD) can be executed correctly
3. There are no errors or exception outputs in both front-end and back-end
4. The system reaches a stable state ready for release to Gitee

Technical implementation requirements:
1. The front-end must use port 3000, and the back-end must use port 8080
2. Implement unified configuration of front-end and back-end ports in the .env configuration file, ensuring that other codes obtain port information by referencing the configuration file
3. Ensure that the run.bat file contains logic to delete all old executable files and compilation products, kill processes occupying ports, and correctly start the front-end port 3000 and back-end port 8080 services
4. All code implementations must be based on the Markdown content requirements in the .trae folder and its subfolders at all levels; no irrelevant code should be generated arbitrarily or unfounded repairs should be made

Verification standards:
- Not only verify whether the page can be opened normally, but also verify the actual function of all function buttons
- Must automatically simulate the complete data entry process
- Must ensure that all front-end and back-end interactions are error-free
- Must ensure the stability of the system under continuous operation