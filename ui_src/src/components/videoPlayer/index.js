import './style.scss';
import React, { useState } from 'react';
import playVideoIcon from '../../assets/images/playVideoIcon.svg';
import ReactPlayer from 'react-player';

const VideoPlayer = (props) => {
    const { url, style } = props;
    const [playState, setPlayState] = useState(false);

    return (
        <ReactPlayer
            style={style}
            controls={true}
            playing={playState}
            light={true}
            playIcon={
                <div onClick={() => setPlayState(true)}>
                    <img alt="play-video-icon" src={playVideoIcon} />
                </div>
            }
            height="250px"
            width="445px"
            url={url}
        ></ReactPlayer>
    );
};

export default VideoPlayer;
