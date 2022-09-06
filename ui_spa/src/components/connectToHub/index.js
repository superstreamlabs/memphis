// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import './style.scss';

import DialogContent from '@material-ui/core/DialogContent';
import { makeStyles } from '@material-ui/core/styles';
import Dialog from '@material-ui/core/Dialog';
import React, { useState } from 'react';
import { Checkbox } from 'antd';

import Close from '../../assets/images/close.svg';
import Input from '../Input';
import Button from '../button';

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
                            <Checkbox checked={formFields.rememberMe} onChange={handelChangeRememberMe} name="checkedG" />
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
