// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

import './style.scss';

import DialogContent from '@material-ui/core/DialogContent';
import { makeStyles } from '@material-ui/core/styles';
import Dialog from '@material-ui/core/Dialog';
import React, { useState } from 'react';

import Close from '../../assets/images/close.svg';
import Input from '../Input';
import Button from '../button';
import CheckboxComponent from '../checkBox';

const useStyles = makeStyles((theme) => ({
    root: {
        flexGrow: 1
    },
    dialogPaper: {
        height: '50vh',
        minHeight: '550px',
        width: '25vw',
        minWidth: '440px',
        borderRadius: '10px',
        padding: '15px'
    }
}));

function ConnectToHub(props) {
    const classes = useStyles();
    const [formFields, setFormFields] = useState({
        username: '',
        password: '',
        rememberMe: true
    });

    const handelChangeUsername = (e) => {
        setFormFields({ ...formFields, username: e.target.value });
    };

    const handelChangePassword = (e) => {
        setFormFields({ ...formFields, password: e.target.value });
    };

    const handelChangeRememberMe = () => {
        setFormFields({ ...formFields, rememberMe: !formFields.rememberMe });
    };

    const clearFormAndClose = () => {
        setFormFields({
            username: '',
            password: '',
            rememberMe: true
        });
        props.closeModeal(false);
    };

    return (
        <Dialog
            open={props.open}
            onClose={(_, reson) => {
                if (reson === 'backdropClick') clearFormAndClose();
                // { props.clickOutside() }
            }}
            classes={{ paper: classes.dialogPaper }}
        >
            <DialogContent className={classes.dialogContent}>
                <div className="connect-to-hub">
                    <div className="connect-to-hub-header">
                        <p>Sign in to hub</p>
                        <img src={Close} alt="close" width="12" height="12" onClick={clearFormAndClose} />
                    </div>
                    <div className="user-password-sectoin">
                        <div className="user-name-input">
                            <p>Username</p>
                            <Input
                                value={formFields.username}
                                placeholder="Type usernmane"
                                type="text"
                                radiusType="semi-round"
                                borderColorType="none"
                                boxShadowsType="gray"
                                colorType="black"
                                backgroundColorType="none"
                                width="21vw"
                                minWidth="360px"
                                height="40px"
                                iconComponent=""
                                onChange={(e) => handelChangeUsername(e)}
                            />
                        </div>
                        <div className="password-input">
                            <p>Password</p>
                            <Input
                                value={formFields.password}
                                placeholder="Type password"
                                type="text"
                                radiusType="semi-round"
                                borderColorType="none"
                                boxShadowsType="gray"
                                colorType="black"
                                backgroundColorType="none"
                                width="21vw"
                                minWidth="360px"
                                height="40px"
                                iconComponent=""
                                onChange={(e) => handelChangePassword(e)}
                            />
                        </div>

                        <span className="remember-me-checkbox" onClick={handelChangeRememberMe}>
                            <CheckboxComponent checked={formFields.rememberMe} id={'checkedG'} onChange={handelChangeRememberMe} name={'checkedG'} />
                            <p>Remember me</p>
                        </span>
                    </div>
                    <div className="sign-in-btn">
                        <Button
                            className="modal-btn"
                            width="100%"
                            height="40px"
                            placeholder="Sign in"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="14px"
                            fontWeight="bold"
                            onClick={() => {}}
                        />
                        <p>Forgot password?</p>
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    );
}
export default ConnectToHub;
