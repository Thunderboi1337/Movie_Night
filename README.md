# Movie_Night
This is an improvement on an old idea and something I did before that required to much manual labor. The Idea was select between about 6 movies. One each from a different genre. One Action, Drama, Comedy, Anime, 3D_Animation, and Weird. And get a link to these movie trailer and an about page to read about them.


## Features

- **Easy Movie selection**: Find the movie you want to add by simply searching for it and add it



## Usage

Once the web application is running, you can use the following functions:

- **Search**: A Search function to easily find the movie you want to add to your movies night.
- **Display Movie aboutpage**: By simply clicking on the poster for any of your selected or searched movies, you will display the AboutPage, where you can learn more about your selected movie.
- **Search and Switch**: Simply search for a movie you are thinking about adding and search for it, When you found the movie you want to add, select the switch button and hit the genre that the movies is going to represent for your Movie Night.
- **Select a Winner**: TO remember what movie won and what movies you watched last week, you can selet a winner by hitting the winner button.

### Example Usage

Here are some GIFs demonstrating how to use the application:

- **Switching a Movie**:  
  ![Add and Complete Task](static/images/No-Picture-Found.png)

- **About Movie**:  
  ![Display Task](static/images/No-Picture-Found.png)

- **Movies**:  
  ![Movies](static/images/No-Picture-Found.png)

## Code Structure

- **Json File**: The server reads from and writes to a json file located in the root folder to get acesses to current or previously selected movies.
- **In-Memory Management**: Movies are managed in memory during the session and saved during switch.
- **UI States**: The user interface supports different states: main page inlcudes selcted movies for current movie night and previous movie night winner. Aboout page that presents all movie information, name, about, trailers.

## Tools

Golang - for server management and api data colletions.
Html & Css - for interaction and apperance.
JavaScript for enhanced functionaltiy.



## License

Feel free to use and modify this project as you see fit. Enjoy!