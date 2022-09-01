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

import WarningRoundedIcon from '@material-ui/icons/WarningRounded';
import CircularProgress from '@material-ui/core/CircularProgress';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import { makeStyles } from '@material-ui/core/styles';
import CloseIcon from '@material-ui/icons/Close';
import Dialog from '@material-ui/core/Dialog';
import React, { useEffect, useState } from 'react';

import Button from '../button';

const TransitionsModal = (props) => {
    const {
        height,
        width,
        rBtnText,
        lBtnText,
        rBtnDisabled,
        lBtnDisabled,
        header,
        confirm,
        minHeight,
        minWidth,
        progress,
        isLoading,
        warning,
        border,
        open = false,
        hr = true
    } = props;

    const useStyles = makeStyles((theme) => ({
        dialogPaper: {
            display: 'flex',
            // padding: "10px",
            justifyContent: 'center',
            width: width,
            height: height,
            border: border,
            borderRadius: '4px',
            minWidth: minWidth || '703px',
            minHeight: minHeight,
            overflowX: 'hidden',
            overflowY: 'auto',
            position: 'relative'
        },
        dialogPaperConfirm: {
            display: 'flex',
            justifyContent: 'center',
            width: props.width ? props.width : '500px',
            height: props.height ? props.height : '230px',
            minWidth: minWidth || '703px',
            minHeight: minHeight,
            border: border,
            borderRadius: '4px',
            overflowY: 'auto'
        },
        buttonLoader: {
            color: '#f7f7f7',
            marginTop: '5px'
        }
    }));

    const classes = useStyles();

    useEffect(() => {
        function handleEscapeKey(event) {
            if (event.code === 'Escape') {
                props.clickOutside();
            }
        }
        document.addEventListener('keyup', handleEscapeKey);
        return () => document.removeEventListener('keydown', handleEscapeKey);
    }, []);

    return confirm ? (
        <Dialog
            onClose={(_, reson) => {
                if (reson === 'backdropClick') {
                    props.clickOutside();
                }
            }}
            open={open}
            aria-labelledby="form-dialog-title"
            classes={{ paper: classes.dialogPaperConfirm }}
        >
            <DialogContent
                style={{
                    position: 'absolute',
                    width: '100%',
                    height: '100%'
                }}
            >
                <span className="header-container">
                    {warning && (
                        <WarningRoundedIcon
                            style={{
                                fontSize: '30px',
                                color: '#fbbd71',
                                marginRight: '10px'
                            }}
                        />
                    )}
                    <p className="modal-header">{header}</p>
                    <CloseIcon onClick={() => props.clickOutside()} style={{ cursor: 'pointer' }} />
                </span>
                {props.children}
            </DialogContent>
            <DialogActions>
                <div
                    style={{
                        position: 'absolute',
                        bottom: '0px',
                        right: '10px'
                    }}
                >
                    <Button
                        className="modal-btn"
                        width="83px"
                        height="32px"
                        placeholder="Close"
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType={warning ? 'orange' : 'purple'}
                        fontSize="12px"
                        fontWeight="600"
                        onClick={() => {
                            props.rBtnClick();
                            props.clickOutside();
                        }}
                    />
                </div>
            </DialogActions>
        </Dialog>
    ) : (
        <Dialog
            open={open}
            onClose={(_, reson) => {
                if (reson === 'backdropClick') {
                    props.clickOutside();
                }
            }}
            aria-labelledby="form-dialog-title"
            classes={{ paper: classes.dialogPaper }}
        >
            <DialogContent
                style={{
                    position: 'absolute',
                    width: '100%',
                    height: '100%'
                }}
            >
                <span className="header-container">
                    {warning && (
                        <WarningRoundedIcon
                            style={{
                                fontSize: '30px',
                                color: '#fbbd71',
                                marginRight: '10px'
                            }}
                        />
                    )}
                    <p className="modal-header">{header}</p>
                    <CloseIcon onClick={() => props.clickOutside()} style={{ cursor: 'pointer' }} />
                </span>
                {props.children}
            </DialogContent>
            <DialogActions>
                {hr && <hr />}
                <div className="btnContainer">
                    <button className="cancel-button" disabled={lBtnDisabled} onClick={() => props.lBtnClick()}>
                        {lBtnText}
                    </button>
                    <Button
                        className="modal-btn"
                        width="83px"
                        height="32px"
                        placeholder={progress ? <CircularProgress size={20} className={classes.buttonLoader} /> : rBtnText}
                        disabled={rBtnDisabled}
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType={warning ? 'orange' : 'purple'}
                        fontSize="12px"
                        fontWeight="600"
                        isLoading={isLoading}
                        onClick={() => {
                            props.rBtnClick();
                        }}
                    />
                </div>
            </DialogActions>
        </Dialog>
    );
};

export default TransitionsModal;
